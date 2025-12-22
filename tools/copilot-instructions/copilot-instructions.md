# Copilot Instructions - Engineering Supervisor Agent

## Architecture Overview

This is a **Gilead enterprise CI/CD platform** with a **multi-service supervisor agent application**. The repository follows a structured approach separating cloud services with a main Container application implementing a conversational AI agent.

### Key Components

- **Container/**: Main supervisor agent application (FastAPI + React + Kubernetes)
- **Compute/**: Lambda functions and API Gateway configurations  
- **DataEngineering/**: Databricks ETL, Glue jobs, AppFlow pipelines
- **DataScience/**: Databricks ML workloads, SageMaker models
- **GenAI/**: Bedrock agents, knowledge bases, vector search, foundation models
- **terraform/**: Infrastructure as Code (SSI team managed)
- **Orchestration/**: MWAA/Airflow DAGs for workflow management

## Container Application (Main Development Focus)

### Architecture Pattern: Enterprise Multi-Tenant SaaS
```
Container/
├── service/         # FastAPI Python backend with middleware-heavy architecture
│   └── src/botofbots/  # Main package: clients/, services/, routers/, config/
├── ui/             # React 18 + TypeScript + Material-UI frontend  
├── helm/           # Kubernetes deployment manifests per environment
└── Makefile        # **PRIMARY DEVELOPMENT INTERFACE** - use this!
```

### FastAPI Backend (`service/src/botofbots/`)

**Core Pattern**: Middleware-driven with service adapters, structured logging, multi-tenant auth
- **app.py**: FastAPI app with authentication, CSRF, session middleware stack
- **routers/v1/**: RESTful API endpoints (chat.py, apps.py)
- **clients/**: External service integrations (Databricks, AWS with retry logic)
- **services/**: Business logic layer (ChatService, ConversationRecord)  
- **data/models/**: Pydantic models, enums, validation schemas
- **config/**: Environment variables, secrets management, constants

**Authentication Flow** (Critical Pattern):
- **Okta OIDC** → JWT access/refresh tokens → encrypted HttpOnly cookies
- **Databricks workspace token exchange** with per-user encryption
- **Session-based CSRF** protection with X-CSRF-Token header validation
- **Multi-tenant isolation**: `request.state.user` (LoggedInUser), `request.state.auth_token`

**Key Services Architecture**:
- **ChatService**: Conversation management via Databricks AI Gateway
- **ConversationRecord**: Persistent chat history to Unity Catalog tables
- **OIDC Databricks**: Secure token exchange for workspace access per user

### React Frontend (`ui/`)

**Stack**: React 18 + TypeScript + Material-UI + Vite + Vitest
- **engines.node**: ">=22" (strict constraint - newer projects need this)
- **Key deps**: @mui/material, ag-grid-react, axios, react-router-dom
- **Build**: TypeScript strict compilation + Vite HMR + production optimized builds
- **Testing**: Vitest with coverage reporting (no E2E, component-focused)

### Development Workflow (Critical - Use Makefile)

**PRIMARY INTERFACE**: Always use `make` commands - this is the authoritative way to work with this codebase:

```bash
# First-time setup
make check-tools    # Verify Python 3.12+, Node 22+, tmux, Caddy prereqs
make install        # Install all dependencies (UI + Python)

# Development modes (choose based on what you're working on)
make start          # Mock backend (fastest - no external deps, pure UI work)
make start-py       # Python backend + mocked external services (backend dev)
make start-py-real  # Python backend + real Databricks/AWS (integration testing)

# Testing & quality (run before PRs)
make test-ui        # Vitest component tests  
make test-service   # pytest backend tests with mocks
make lint-ui        # ESLint + Prettier
make lint-service   # flake8 Python linting
make validate-all   # ALL checks - required before PR/deploy

# GitHub operations (safety-first pattern)
make create-pr      # Dry-run PR creation preview (safe)
make create-pr-apply # Create PR with full validation + GitHub Copilot description
make deploy-dev     # Dry-run deployment preview (safe)  
make deploy-dev-apply # Deploy with validation to dev environment
```

**Safety Pattern**: All destructive operations default to dry-run mode. Use `-apply` suffix after reviewing dry-run output.

**Port Layout** (memorize this):
- **3006**: Caddy HTTPS proxy (**USE THIS** - https://localhost:3006)
- **3000**: React dev server (direct access may have CORS issues)
- **8000**: Python FastAPI backend
- **8001**: Mock backend server

## Environment Configuration Patterns

### Backend Environment Variables (Required for `start-py-real`)
```bash
# Okta OIDC (multi-tenant auth)
OKTA_ISSUER=https://gsso.gilead.com
OKTA_CLIENT_ID=xxxxx  
OKTA_CLIENT_SECRET=xxxxx

# Databricks (per-environment workspace)
DATABRICKS_HOST=https://dbc-xxxxx.cloud.databricks.com
DATABRICKS_TOKEN=xxxxx
DATABRICKS_TOKEN_ENCRYPTION_KEY=xxxxx  # Fernet encryption key
UNITY_CATALOG_NAME=engineering-supervisor-agent-dev  
UNITY_SCHEMA_NAME=conversation_ui

# Security
JWT_SECRET_KEY=xxxxx
SESSION_SECRET=xxxxx
FRONTEND_URL=https://localhost:3000/chat  # CSRF origin validation
```

## Deployment Architecture (GitHub Actions + EKS)

### Trigger Pattern
**PR Comments**: `deploy --target=Container` triggers GitHub Actions workflows
- Builds Docker images with multi-stage builds
- Deploys via Helm charts to EKS clusters 
- Environment-specific deployment (dev/tst/val/prd via tfvars)

### Infrastructure Pattern  
- **EKS Kubernetes**: Managed by SSI team via Terraform
- **Unity Catalog**: Per-environment Databricks schema isolation  
- **Okta OIDC**: Enterprise SSO with SAML metadata
- **AWS Secrets Manager**: Shared Databricks credentials with encryption

## Development Patterns (Follow These)

### Authentication Architecture (Current - Middleware Pattern)
```python
# Middleware-based auth (current implementation)
@app.middleware("http")
async def authenticate(request: Request, call_next):
    # Token validation, refresh, workspace token exchange
    request.state.user = LoggedInUser(...)
    request.state.auth_token = encrypted_workspace_token
    
# Route dependencies
async def get_current_user(request: Request) -> LoggedInUser:
    return request.state.user

@router.post("/endpoint")  
async def endpoint(user: LoggedInUser = Depends(get_current_user)):
    # Business logic with authenticated user context
```

### Multi-Tenant Databricks Pattern (Critical)
- **Per-user workspace tokens** stored in encrypted cookies
- **Automatic token refresh** in middleware with exponential backoff
- **Unity Catalog isolation** per user/environment schema
- **Workspace-specific endpoints** via token exchange

### File Organization Standards
```
service/src/botofbots/
├── routers/v1/          # API routes only - no business logic
├── services/            # Business logic layer - core application logic  
├── clients/             # External service adapters with error handling
├── config/              # Environment/secrets/constants - centralized config
├── data/models/         # Pydantic schemas, enums, validation
└── exceptions.py        # Custom exception hierarchy
```

### Error Handling Pattern
```python
from botofbots.structured_logger import get_logger
logger = get_logger(__name__)

try:
    result = external_service_call()
except ExternalServiceError as e:
    logger.error("Service integration failed", 
                error=str(e), 
                service="databricks", 
                user_id=user.id,
                workspace_url=workspace_host)
    raise HTTPException(status_code=503, detail="External service unavailable")
```

### Configuration Management (Strict Pattern)
- **Environment vars**: `config/environment_variables.py` only
- **Secrets**: `config/secrets.py` with encryption/decryption utils  
- **Constants**: `config/constants.py` for app-wide constants
- **NO hardcoded values** in business logic - use config injection

## CI/CD Comment Commands (GitHub PR Integration)

Trigger deployments via PR comments:
- `deploy ws` - Deploy ALL workspace components  
- `deploy --target=Container` - Deploy container application only
- `deploy --target=DataEngineering/Glue` - Deploy specific pipeline component
- `deploy --target=GenAI/BedrockAgent` - Deploy AI components

## Testing Strategy (Component-Focused)

### UI Testing (`make test-ui`)
- **Vitest** component tests - no E2E, focus on component behavior
- **Mock strategy**: API responses mocked via MSW patterns
- **Coverage**: Component logic, not integration flows

### Backend Testing (`make test-service`)  
- **pytest** with unit + integration test separation
- **Mock strategy**: `tests/testbotofbots/localmock/` provides realistic external service responses
- **Database mocking**: Unity Catalog operations mocked with realistic schema

### Code Quality Standards
- **Python**: flake8 linting, structured logging with correlation IDs
- **TypeScript**: ESLint strict, Material-UI component patterns
- **Security**: CodeQL scanning, dependency vulnerability scanning, CSRF validation

## Technical Debt & Improvement Areas

**Authentication Refactoring**: See `Container/AUTHENTICATION_ADAPTER_DESIGN.md` - planned migration from middleware to adapter pattern for better testability and separation of concerns.

**Preserve These Patterns**: The codebase demonstrates mature enterprise patterns around:
- Structured logging with correlation tracking
- Multi-tenant configuration management  
- Security-first authentication with token encryption
- Environment-specific deployment isolation

Always prioritize these established patterns when making changes - they reflect hard-learned lessons in enterprise AI application development.