# Python Production Dockerfile
# Supports Flask, Django, FastAPI with gunicorn

FROM python:3.12-slim AS builder
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

FROM python:3.12-slim
WORKDIR /app

RUN addgroup --system appgroup && adduser --system --group appuser

COPY --from=builder /install /usr/local
COPY . .

RUN python -m compileall . 2>/dev/null || true

USER appuser
EXPOSE 8000
CMD ["gunicorn", "--bind", "0.0.0.0:8000", "--workers", "4", "app:app"]
