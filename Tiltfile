# -*- mode: Python -*-

docker_build("rainbow-road", ".")
k8s_yaml('k8s/deployment.yaml')

local_resource(
    name='github-token',
    cmd='kubectl create secret generic github-token --from-literal=GITHUB_TOKEN=$GITHUB_TOKEN --dry-run=client -o yaml | kubectl apply -f -'
)

k8s_resource('rainbow-road-api', port_forwards=9999)