# HUB - Caching proxy server

Config example:
```yaml
dir: _data
server:
  pypi:
    pypi.org: https://pypi.org/simple
  galaxy:
    ansible:
      url: https://galaxy.ansible.com
    test:
      dir: _galaxy
  static:
    github: https://github.com
    k8s: https://dl.k8s.io
    get_helm: https://get.helm.sh
```
