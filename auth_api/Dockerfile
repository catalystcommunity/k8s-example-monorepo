FROM python:3.13-slim-bookworm

WORKDIR /app

# The installer requires curl (and certificates) to download the release archive
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates

# Download the latest installer
ADD https://astral.sh/uv/install.sh /uv-installer.sh

# Run the installer then remove it
RUN sh /uv-installer.sh && rm /uv-installer.sh

RUN apt-get clean -y && apt-get remove -y --autoremove curl

# Ensure the installed binary is on the `PATH`
ENV PATH="/root/.local/bin/:$PATH"

# Copy requirements first for better caching
COPY pyproject.toml uv.lock /app/

# Sync the project into a new environment, using the frozen lockfile
RUN uv sync --frozen

# Copy the project into the image
ADD . /app

ENTRYPOINT ["uv", "run", "main.py"]
CMD ["uv", "run", "main.py"]
