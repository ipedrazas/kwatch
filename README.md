# KWatch - Kubernetes Watcher


Kwatch watches a kubernetes endpoint and notifies changes using a hook. It uses HTTP to connect with the API Server, no kubectl or binaries, just good old http requests.

Authentication is done via Basic Auth for now, but soon Token based auth will be added.

