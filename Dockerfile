FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .

RUN pip install -r requirements.txt

COPY app/ .

COPY migrations/ /migrations/

COPY scripts/ ./scripts/

RUN chmod +x ./scripts/*.sh

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]