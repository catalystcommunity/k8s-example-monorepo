FROM uv:0.4.15

WORKDIR /workspace

COPY ./pyproject.toml /workspace/pyproject.toml

RUN uv sync --frozen

COPY ./ /workspace/

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
