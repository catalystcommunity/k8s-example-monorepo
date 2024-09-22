FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim

WORKDIR /workspace

COPY ./pyproject.toml /workspace/pyproject.toml
COPY ./uv.lock /workspace/uv.lock

RUN uv sync --frozen --no-install-project --no-dev

COPY ./ /workspace/

# Place executables in the environment at the front of the path
ENV PATH="/workspace/.venv/bin:$PATH"

# Reset the entrypoint, don't invoke `uv`
ENTRYPOINT []

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
