import uvicorn

from auth.config import Config
from auth.handlers import App


def main():
    uvicorn.run(App, port=Config.port)


if __name__ == '__main__':
    main()
