# -*- mode: Python -*-

docker_build("rainbow-road", ".")
k8s_yaml('k8s/deployment.yaml')
k8s_resource('rainbow-road-api', port_forwards=9999)