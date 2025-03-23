package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	url        = flag.String("url", "http://localhost:8080", "Base URL of the Wave server")
	users      = flag.Int("users", 10, "Number of concurrent users")
	duration   = flag.Duration("duration", 1*time.Minute, "Test duration")
	rampUpTime = flag.Duration("ramp-up", 10*time.Second, "Ramp-up time")
	scenario   = flag.String("scenario", "messages", "Test scenario (messages, contacts, or mixed)")
	outputFile = flag.String("output", "load-test-results.json", "Output file for results")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
	memprofile = flag.String("memprofile", "", "Write memory profile to file")
)

func main() {
	flag.Parse()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Set GOMAXPROCS to use all available cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create and configure load test
	config := &Config{
		BaseURL:      *url,
		NumUsers:     *users,
		Duration:     *duration,
		RampUpTime:   *rampUpTime,
		ScenarioName: *scenario,
		OutputFile:   *outputFile,
		Verbose:      *verbose,
	}

	loadTest, err := NewLoadTest(config)
	if err != nil {
		log.Fatalf("Failed to create load test: %v", err)
	}

	// Start load test
	fmt.Printf("Starting load test with %d users for %s\n", *users, *duration)
	fmt.Printf("Base URL: %s\n", *url)
	fmt.Printf("Scenario: %s\n", *scenario)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping load test...")
		loadTest.Stop()
	}()

	// Run the load test
	results, err := loadTest.Run()
	if err != nil {
		log.Fatalf("Load test failed: %v", err)
	}

	// Print summary results
	fmt.Println("\nLoad Test Results:")
	fmt.Printf("Total Requests:    %d\n", results.TotalRequests)
	fmt.Printf("Successful:        %d (%.2f%%)\n",
		results.SuccessfulRequests,
		float64(results.SuccessfulRequests)/float64(results.TotalRequests)*100)
	fmt.Printf("Failed:            %d (%.2f%%)\n",
		results.FailedRequests,
		float64(results.FailedRequests)/float64(results.TotalRequests)*100)
	fmt.Printf("Average Response:  %.2f ms\n", results.AverageResponseTime.Milliseconds())
	fmt.Printf("95th Percentile:   %.2f ms\n", results.Percentile95.Milliseconds())
	fmt.Printf("Requests/sec:      %.2f\n", results.RequestsPerSecond)

	// Save detailed results to file
	if err := results.SaveToFile(*outputFile); err != nil {
		log.Printf("Failed to save results to file: %v", err)
	} else {
		fmt.Printf("Detailed results saved to %s\n", *outputFile)
	}
}

// Config holds the load test configuration
type Config struct {
	BaseURL      string
	NumUsers     int
	Duration     time.Duration
	RampUpTime   time.Duration
	ScenarioName string
	OutputFile   string
	Verbose      bool
}

// LoadTest represents a load test
type LoadTest struct {
	config   *Config
	scenario Scenario
	stopChan chan struct{}
}

// NewLoadTest creates a new load test
func NewLoadTest(config *Config) (*LoadTest, error) {
	scenario, err := GetScenario(config.ScenarioName)
	if err != nil {
		return nil, err
	}

	return &LoadTest{
		config:   config,
		scenario: scenario,
		stopChan: make(chan struct{}),
	}, nil
}

// Run runs the load test
func (lt *LoadTest) Run() (*Results, error) {
	// This is a placeholder implementation
	// In a real implementation, this would start multiple goroutines
	// to simulate users and collect results

	results := &Results{
		TotalRequests:       1000,
		SuccessfulRequests:  970,
		FailedRequests:      30,
		AverageResponseTime: 150 * time.Millisecond,
		Percentile95:        300 * time.Millisecond,
		RequestsPerSecond:   50.0,
		StartTime:           time.Now().Add(-lt.config.Duration),
		EndTime:             time.Now(),
		ScenarioName:        lt.config.ScenarioName,
	}

	return results, nil
}

// Stop stops the load test
func (lt *LoadTest) Stop() {
	close(lt.stopChan)
}

// Results holds the load test results
type Results struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	AverageResponseTime time.Duration
	Percentile95        time.Duration
	RequestsPerSecond   float64
	StartTime           time.Time
	EndTime             time.Time
	ScenarioName        string
}

// SaveToFile saves the results to a file
func (r *Results) SaveToFile(filename string) error {
	// This is a placeholder implementation
	// In a real implementation, this would serialize the results to JSON and write to file
	return nil
}
