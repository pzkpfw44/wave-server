// Modify the import paths to use github.com/pzkpfw44/wave-server
import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/pzkpfw44/wave-server/internal/api"
	"github.com/pzkpfw44/wave-server/internal/api/handlers"
	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/security"
	"github.com/pzkpfw44/wave-server/internal/service"
	"github.com/pzkpfw44/wave-server/pkg/health"
)