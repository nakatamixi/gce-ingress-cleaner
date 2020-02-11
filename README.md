# gce-ingress-cleaner

GKE Ingress is implemented with GCE resources. 
When GKE Ingress deleted, the related GCE resources will not be deleted.
This tool print  delete commands using gcloud for deleting related GCE resources.

# Usage
```
go run cmd/ingress-gce-cleaner/main.go -n $namespace -i $ingress
```
you need kubectl and permission to read target ingress.
