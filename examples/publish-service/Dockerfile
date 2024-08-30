# echo-service/Dockerfile
FROM python:3.9-slim

WORKDIR /app

COPY main.py .

RUN pip install --no-cache-dir google-cloud-pubsub

CMD ["python", "main.py"]
