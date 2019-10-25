sudo docker build -f task-server/Dockerfile .  -t task-server
sudo docker build . -t api-server
sudo docker-compose up