package yugabytedb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// TableShardConfig defines sharding configuration for a table
type TableShardConfig struct {
	TableName      string
	ShardingKey    string
	NumShards      int
	ReplicaPlaces  int
	ColocatedWith  string
	PartitionBound string
}

// SetupSharding configures table sharding for YugabyteDB
func SetupSharding(ctx context.Context, pool *pgxpool.Pool, configs []TableShardConfig, logger *zap.Logger) error {
	for _, config := range configs {
		// Check if table already exists
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = $1
			)
		`, config.TableName).Scan(&exists)
		
		if err != nil {
			return fmt.Errorf("failed to check if table exists: %w", err)
		}
		
		if !exists {
			logger.Warn("Table does not exist, skipping sharding configuration", 
				zap.String("table", config.TableName))
			continue
		}
		
		// Configure sharding
		if config.ShardingKey != "" {
			// For YugabyteDB, we use the CREATE INDEX ... WITH (split_at) syntax to control sharding
			indexName := fmt.Sprintf("idx_%s_shard", config.TableName)
			query := fmt.Sprintf(`
				CREATE INDEX IF NOT EXISTS %s ON %s(%s)
				WITH (split_at = %d)
			`, indexName, config.TableName, config.ShardingKey, config.NumShards)
			
			_, err := pool.Exec(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to create sharding index for table %s: %w", config.TableName, err)
			}
			
			logger.Info("Configured sharding for table", 
				zap.String("table", config.TableName),
				zap.String("key", config.ShardingKey),
				zap.Int("shards", config.NumShards))
		}
		
		// Configure colocation if specified
		if config.ColocatedWith != "" {
			query := fmt.Sprintf(`
				ALTER TABLE %s SET (colocated_with = '%s')
			`, config.TableName, config.ColocatedWith)
			
			_, err := pool.Exec(ctx, query)
			if err != nil {
				logger.Warn("Failed to set colocation", 
					zap.String("table", config.TableName),
					zap.String("colocated_with", config.ColocatedWith),
					zap.Error(err))
				// Continue with other tables even if this fails
			} else {
				logger.Info("Set colocation for table", 
					zap.String("table", config.TableName),
					zap.String("colocated_with", config.ColocatedWith))
			}
		}
	}
	
	return nil
}

// ApplyMultiNodeConfig applies configuration optimized for multi-node deployment
func ApplyMultiNodeConfig(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) error {
	// Tables that contain conversations between users should be sharded by conversation ID
	tableConfigs := []TableShardConfig{
		{
			TableName:     "messages",
			ShardingKey:   `(CASE WHEN sender_pubkey < recipient_pubkey THEN sender_pubkey || recipient_pubkey ELSE recipient_pubkey || sender_pubkey END)`,
			NumShards:     16, // Adjust based on cluster size
			ReplicaPlaces: 3,  // RF=3 for high availability
		},
		{
			TableName:     "contacts",
			ShardingKey:   "user_id",
			NumShards:     8,
			ReplicaPlaces: 3,
		},
		{
			TableName:     "users",
			ShardingKey:   "user_id",
			NumShards:     8,
			ReplicaPlaces: 3,
		},
		{
			TableName:     "tokens",
			ShardingKey:   "user_id",
			NumShards:     8,
			ReplicaPlaces: 3,
			ColocatedWith: "users", // Colocate tokens with users for faster access
		},
	}
	
	return SetupSharding(ctx, pool, tableConfigs, logger)
}

// GetClusterStatus gets the status of the YugabyteDB cluster
func GetClusterStatus(ctx context.Context, pool *pgxpool.Pool) (map[string]interface{}, error) {
	// This is a YugabyteDB-specific query to get cluster status
	const query = `SELECT * FROM yb_servers()`
	
	rows, err := pool.Query(ctx, query)
	if err != nil {
		// If the function doesn't exist, we're not connected to a YugabyteDB cluster
		if strings.Contains(err.Error(), "function yb_servers() does not exist") {
			return map[string]interface{}{
				"is_yugabyte": false,
				"error":       "Connected to PostgreSQL, not YugabyteDB",
			}, nil
		}
		return nil, fmt.Errorf("failed to query cluster status: %w", err)
	}
	defer rows.Close()
	
	servers := make([]map[string]interface{}, 0)
	
	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}
	
	// Scan rows
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		server := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			server[col] = val
		}
		
		servers = append(servers, server)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return map[string]interface{}{
		"is_yugabyte": true,
		"servers":     servers,
		"node_count":  len(servers),
	}, nil
}