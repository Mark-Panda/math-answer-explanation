# 后端（在项目根目录）
export GOMATH_MODELS_CONFIG=config/models.yaml  # 可选
go run ./cmd/gomath

# 前端（另开终端）
cd web && npm run dev