# deploy-compose

A utiliy to deploy services defined in a compose file to swarm-mode clusters.

## Usage

```
  deploy-compose [options] [COMMAND]
  deploy-compose -h|--help

Options:
  -f string
    	Specify an alternate compose file (default "docker-compose.yml")
  -p string
    	Specify an alternate project name (default: directory name)

Commands:
  up                 Create and start services, networks, and volumes
  down               Stop and remove containers, networks, and volumes
```
