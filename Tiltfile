K8S_NAMESPACE = "redalert-local"

load('ext://deployment', 'deployment_create')
load('ext://namespace', 'namespace_create')

allow_k8s_contexts('docker-desktop')

def redalert(namespace):
       docker_build(
              'redalert',
              '.',
              dockerfile='./Dockerfile',
              platform='linux/arm64')

       k8s_yaml(helm(
              './charts/redalert',
              name='redalert-local',
              namespace=namespace,
              set=[
                     'image.pullPolicy=IfNotPresent',
                     'image.repository=redalert',
                     'image.tag=latest',
              ]
              ))
       k8s_resource(
              workload='redalert-local',
              new_name='www',
              port_forwards=[port_forward(8888, 8888, '/')],
              links=[
                     link('http://127.0.0.1:8888', 'web')],
              labels=['backend'])

namespace_create(K8S_NAMESPACE)
redalert(K8S_NAMESPACE)
