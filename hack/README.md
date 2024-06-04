# hack

## Local webhook development

```bash
ngrok http https://localhost:9443
```

```bash
export WEBHOOK_BASE_URL=https://xyz.ngrok-free.app

hack/webhook-install.sh

make run
```

## Percona cluster

```bash
kubectl create secret generic percona \
    --from-literal=root=$(pwgen 20 1) \
    --from-literal=xtrabackup=$(pwgen 20 1) \
    --from-literal=monitor=$(pwgen 20 1) \
    --from-literal=proxyadmin=$(pwgen 20 1) \
    --from-literal=operator=$(pwgen 20 1) \
    --from-literal=replication=$(pwgen 20 1)
```
