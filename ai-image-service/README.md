# AI Image Service

FastAPI adapter used by the Go API to generate admin images with MiniMax image generation.

## Local Run

```powershell
cd ai-image-service
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
$env:MINIMAX_API_KEY="your-key"
uvicorn app.main:app --host 0.0.0.0 --port 8090
```

Set the Go API environment variable:

```powershell
$env:AI_IMAGE_SERVICE_URL="http://127.0.0.1:8090"
```

## Endpoint

`POST /v1/images/generate`

```json
{
  "prompt": "清爽水果分类封面，苹果、橙子、蓝莓，明亮电商风格",
  "usage": "category-cover",
  "width": 1024,
  "height": 1024,
  "aspect_ratio": "1:1"
}
```

The response is normalized to either `image_base64` or `image_url`; the Go API stores the final image in the configured object storage.

MiniMax image generation uses `POST /v1/image_generation`. The default image model is `image-01`; `MiniMax-M2.7` is a text-chat model and is not accepted by the image generation endpoint.
