# calculator-api

API HTTP em Go (só biblioteca padrão) com as quatro operações básicas, empacotada para Kubernetes.

| Endpoint | Exemplo | Resposta |
| --- | --- | --- |
| `GET /api/sum` | `/api/sum?term_one=4&term_two=1` | `{"result":5}` |
| `GET /api/sub` | `/api/sub?term_one=4&term_two=1` | `{"result":3}` |
| `GET /api/mul` | `/api/mul?term_one=4&term_two=1` | `{"result":4}` |
| `GET /api/div` | `/api/div?term_one=4&term_two=1` | `{"result":4}` |
| `GET /healthz` | | `{"status":"ok"}` |

- Porta `8000`.
- Termos e resultado são **inteiros**, como o enunciado define (`<int:result>`). A divisão é inteira e trunca em direção a zero: `5/2 = 2`.
- Erros respondem `400` com `{"error": "..."}`: termo ausente ou não inteiro, e divisão por zero.
- A API só é acessível de dentro do cluster: o Service é `ClusterIP`, sem NodePort, LoadBalancer ou Ingress.

## Pré-requisitos

`kubectl` apontando para o cluster e `make`. Para construir a imagem, também `docker`; para os testes, `go` 1.26.

## Deploy

**Opção 1: imagem publicada (não precisa de Docker)**

```bash
kubectl apply -f k8s/
kubectl -n calculator rollout status deployment/calculator-api
```

Usa `ghcr.io/raulmoura157/calculator-api:latest`, que o CI publica a cada push na `main`.

**Opção 2: construir a imagem a partir do código** (é a opção a usar depois de alterar a API)

```bash
# qualquer cluster: build + push para um registry que o cluster alcance + deploy
make deploy REGISTRY=<seu-registry>          # ex.: REGISTRY=docker.io/usuario

# cluster kind: carrega a imagem direto no cluster, sem registry
make kind-deploy KIND_CLUSTER=<nome-do-cluster>
```

Cada build recebe uma tag única (timestamp), então o deploy sempre gera um rollout novo, mesmo com mudanças não commitadas. O rollout é gradual (`maxUnavailable: 0`) e só avança quando o `readinessProbe` em `/healthz` responde.

## Validar

Do lado de dentro do cluster, com um pod descartável:

```bash
make smoke
# ou manualmente:
kubectl -n calculator run curl --rm -i --restart=Never --image=curlimages/curl:8.11.1 -- \
  curl -s "http://calculator-api:8000/api/sub?term_one=4&term_two=1"
```

De outro namespace, o endereço é `http://calculator-api.calculator.svc.cluster.local:8000`.

## Alterar a API e fazer um novo deploy

1. Edite `main.go`. Cada operação é uma entrada no mapa `operations`, e uma operação nova vira um endpoint `/api/<nome>` automaticamente.
2. Rode `make test`.
3. Faça o deploy com `make deploy REGISTRY=...` ou `make kind-deploy`.

## Remover

```bash
make undeploy        # equivale a: kubectl delete -f k8s/
```

Isso apaga o namespace `calculator` e tudo o que está dentro dele.

## Testes e cobertura

```bash
make test            # testes + percentual de cobertura
make cover           # gera coverage.out e coverage.html (relatório navegável)
```

O CI (`.github/workflows/ci.yml`) roda `go vet` e os testes, publica `coverage.html` como artefato e, na `main`, publica a imagem no GHCR.

## Estrutura

```
main.go            API (handlers + healthcheck)
main_test.go       testes de tabela via httptest
Dockerfile         build multi-stage, imagem final distroless, usuário não-root
k8s/               00-namespace, 01-service (ClusterIP), 02-deployment (probes, resources, securityContext)
Makefile           test, cover, build, deploy, kind-deploy, smoke, undeploy
```
