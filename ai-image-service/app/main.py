import os
from typing import Any

import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field


class GenerateImageRequest(BaseModel):
    prompt: str = Field(min_length=1, max_length=1200)
    usage: str = "common"
    width: int = 1024
    height: int = 1024
    aspect_ratio: str = "1:1"


class GenerateImageResponse(BaseModel):
    image_base64: str | None = None
    image_url: str | None = None
    content_type: str = "image/jpeg"
    model: str
    revised_prompt: str = ""


app = FastAPI(title="Go Shoppings AI Image Service", version="0.1.0")

ASPECT_RATIO_VALUES = {
    "1:1": 1,
    "16:9": 16 / 9,
    "4:3": 4 / 3,
    "3:2": 3 / 2,
    "2:3": 2 / 3,
    "3:4": 3 / 4,
    "9:16": 9 / 16,
    "21:9": 21 / 9,
}


def env(name: str, default: str = "") -> str:
    return os.getenv(name, default).strip()


def usage_instruction(
    usage: str,
    width: int,
    height: int,
    aspect_ratio: str,
) -> str:
    labels = {
        "category-cover": "微信商城商品分类封面图",
        "storefront-banner": "微信商城首页 Banner 横幅图",
        "product-cover": "微信商城商品封面图",
        "product-gallery": "微信商城商品详情图",
        "brand-logo": "品牌 Logo",
        "platform-logo": "平台 Logo",
    }
    label = labels.get(usage, "微信商城运营图片")
    return (
        f"用途：{label}。尺寸建议：{width}x{height}，比例 {aspect_ratio}。"
        "画面清晰、商业摄影质感、适合电商运营，不要出现多余文字、水印、二维码、边框。"
    )


def round_to_multiple(value: float, multiple: int) -> int:
    return max(multiple, int(round(value / multiple)) * multiple)


def normalize_dimensions(width: int, height: int) -> tuple[int, int]:
    target_width = max(width or 1024, 1)
    target_height = max(height or 1024, 1)
    scale = max(512 / target_width, 512 / target_height, 1)
    if target_width * scale > 2048 or target_height * scale > 2048:
        scale = min(2048 / target_width, 2048 / target_height)

    normalized_width = round_to_multiple(target_width * scale, 8)
    normalized_height = round_to_multiple(target_height * scale, 8)
    return (
        min(2048, max(512, normalized_width)),
        min(2048, max(512, normalized_height)),
    )


def ratio_to_float(value: str, width: int, height: int) -> float:
    if ":" in value:
        left, right = value.split(":", 1)
        try:
            return float(left) / float(right)
        except (TypeError, ValueError, ZeroDivisionError):
            pass
    if height > 0:
        return width / height
    return 1


def nearest_aspect_ratio(value: str, width: int, height: int) -> str:
    if value in ASPECT_RATIO_VALUES:
        return value
    ratio = ratio_to_float(value, width, height)
    return min(
        ASPECT_RATIO_VALUES,
        key=lambda item: abs(ASPECT_RATIO_VALUES[item] - ratio),
    )


def minimax_payload(req: GenerateImageRequest) -> dict[str, Any]:
    model = env("MINIMAX_IMAGE_MODEL", "image-01")
    instruction = usage_instruction(
        req.usage,
        req.width,
        req.height,
        req.aspect_ratio,
    )
    prompt = f"{req.prompt}\n\n{instruction}"
    payload: dict[str, Any] = {
        "model": model,
        "prompt": prompt,
        "n": 1,
        "response_format": "url",
        "prompt_optimizer": True,
        "aigc_watermark": False,
    }
    if model == "image-01-live":
        payload["aspect_ratio"] = nearest_aspect_ratio(
            req.aspect_ratio,
            req.width,
            req.height,
        )
    else:
        width, height = normalize_dimensions(req.width, req.height)
        payload["width"] = width
        payload["height"] = height
    return payload


def first_string(value: Any) -> str | None:
    if isinstance(value, str) and value:
        return value
    if isinstance(value, list):
        for item in value:
            if isinstance(item, str) and item:
                return item
    return None


def minimax_error(payload: dict[str, Any]) -> str:
    base_resp = payload.get("base_resp")
    if not isinstance(base_resp, dict):
        return ""
    status_code = base_resp.get("status_code", 0)
    status_msg = base_resp.get("status_msg") or "unknown error"
    if status_code in (0, "0", None):
        return ""
    return f"Minimax error {status_code}: {status_msg}"


def pick_first_image(
    payload: dict[str, Any],
) -> tuple[str | None, str | None, str]:
    data = payload.get("data")
    if isinstance(data, dict):
        image_base64 = first_string(
            data.get("image_base64")
            or data.get("image_base64s")
            or data.get("base64")
            or data.get("base64s")
        )
        image_url = first_string(
            data.get("image_url")
            or data.get("image_urls")
            or data.get("url")
            or data.get("urls")
        )
        if image_base64 or image_url:
            return image_base64, image_url, payload.get("revised_prompt") or ""

    if isinstance(data, list) and data:
        first = data[0] or {}
        if isinstance(first, dict):
            return (
                first.get("b64_json") or first.get("base64"),
                first.get("url"),
                first.get("revised_prompt") or "",
            )

    if payload.get("image_base64") or payload.get("image_url"):
        return (
            payload.get("image_base64"),
            payload.get("image_url"),
            payload.get("revised_prompt") or "",
        )

    result = payload.get("result")
    if isinstance(result, dict):
        return (
            result.get("image_base64") or result.get("base64"),
            result.get("image_url") or result.get("url"),
            result.get("revised_prompt") or "",
        )

    return None, None, ""


@app.get("/healthz")
async def healthz() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/v1/images/generate", response_model=GenerateImageResponse)
async def generate_image(req: GenerateImageRequest) -> GenerateImageResponse:
    api_key = env("MINIMAX_API_KEY")
    if not api_key:
        raise HTTPException(
            status_code=500,
            detail="MINIMAX_API_KEY is not configured",
        )

    base_url = env(
        "MINIMAX_API_BASE_URL",
        "https://api.minimaxi.com/v1",
    ).rstrip("/")
    endpoint = env("MINIMAX_IMAGE_ENDPOINT", "/image_generation")
    timeout = float(env("MINIMAX_TIMEOUT_SECONDS", "120"))
    model = env("MINIMAX_IMAGE_MODEL", "image-01")

    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    }
    async with httpx.AsyncClient(timeout=timeout) as client:
        resp = await client.post(
            f"{base_url}{endpoint}",
            headers=headers,
            json=minimax_payload(req),
        )
    if resp.status_code >= 400:
        raise HTTPException(
            status_code=502,
            detail=f"Minimax returned {resp.status_code}: {resp.text}",
        )

    payload = resp.json()
    err_msg = minimax_error(payload)
    if err_msg:
        raise HTTPException(status_code=502, detail=err_msg)

    image_base64, image_url, revised_prompt = pick_first_image(payload)
    if not image_base64 and not image_url:
        raise HTTPException(
            status_code=502,
            detail="Minimax response did not include an image",
        )

    return GenerateImageResponse(
        image_base64=image_base64,
        image_url=image_url,
        content_type=payload.get("content_type") or "image/jpeg",
        model=payload.get("model") or model,
        revised_prompt=revised_prompt,
    )
