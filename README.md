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

`kubectl` apontando para o cluster. Para construir a imagem, também `docker`; para os testes, `go` 1.26.

## Deploy

```bash
kubectl apply -f k8s/
kubectl -n calculator rollout status deployment/calculator-api
```

Usa a imagem `ghcr.io/raulmoura157/calculator-api:latest`, que o CI publica a cada push na `main`. Os arquivos em `k8s/` têm prefixo numérico para o Namespace ser aplicado antes dos outros recursos.

## Validar

De dentro do cluster, com um pod descartável:

```bash
kubectl -n calculator run curl --rm -i --restart=Never --image=curlimages/curl:8.11.1 -- \
  curl -s "http://calculator-api:8000/api/sub?term_one=4&term_two=1"
# {"result":3}
```

De outro namespace, o endereço é `http://calculator-api.calculator.svc.cluster.local:8000`.

Para chamar da sua máquina, abra um túnel:

```bash
kubectl -n calculator port-forward svc/calculator-api 8000:8000
curl "localhost:8000/api/sum?term_one=4&term_two=1"
```

## Alterar a API e fazer um novo deploy

1. Edite `main.go`. Cada operação é uma entrada no mapa `operations`, e uma operação nova vira um endpoint `/api/<nome>` automaticamente.
2. Rode os testes: `go test ./...`
3. Construa a imagem com uma **tag nova** e publique num registry que o cluster alcance:

   ```bash
   IMAGE=<seu-registry>/calculator-api:v2     # ex.: docker.io/usuario/calculator-api:v2
   docker build -t $IMAGE .
   docker push $IMAGE
   ```

   Num cluster [kind](https://kind.sigs.k8s.io/), troque o `docker push` por `kind load docker-image $IMAGE --name <cluster>`.

4. Atualize o Deployment e acompanhe o rollout:

   ```bash
   kubectl -n calculator set image deployment/calculator-api calculator-api=$IMAGE
   kubectl -n calculator rollout status deployment/calculator-api
   ```

O rollout é gradual (`maxUnavailable: 0`) e só avança quando o `readinessProbe` em `/healthz` responde. Use sempre uma tag nova: repetir a mesma tag não gera rollout.

## Remover

```bash
kubectl delete -f k8s/
```

Isso apaga o namespace `calculator` e tudo o que está dentro dele.

## Testes e cobertura

```bash
go test -cover ./...                          # testes + percentual de cobertura
go test -coverprofile=coverage.out ./...      # relatório detalhado
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

O CI (`.github/workflows/ci.yml`) roda `go vet` e os testes, publica `coverage.html` como artefato e, na `main`, publica a imagem no GHCR.

## Estrutura

```
main.go            API (handlers + healthcheck)
main_test.go       testes de tabela via httptest
Dockerfile         build multi-stage, imagem final distroless, usuário não-root
k8s/               00-namespace, 01-service (ClusterIP), 02-deployment (probes, resources, securityContext)
```
