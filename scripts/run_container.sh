docker run -d -p 127.0.0.1:8000:8000 --name workspace-projects --network workspace --restart unless-stopped --env-file .env workspace-projects
