# calculator-api

Solução do teste para a vaga de SRE: uma **aplicação web** (API em Go) com as quatro operações básicas, e o **deploy** dela em Kubernetes.

## Passo a passo

Precisa só de `kubectl` apontando para o cluster. Todos os comandos rodam a partir da raiz do repositório.

**1. Fazer o deploy**

```bash
kubectl apply -f k8s/
kubectl -n calculator rollout status deployment/calculator-api
```

**2. Testar a API** (de dentro do cluster, com um pod temporário)

```bash
kubectl -n calculator run curl --rm -i --restart=Never --image=curlimages/curl:8.11.1 -- \
  curl -s "http://calculator-api:8000/api/sub?term_one=4&term_two=1"
```

Resposta esperada: `{"result":3}`

**3. Alterar a API e fazer um novo deploy** (precisa também de `docker`)

Edite o `main.go` e rode:

```bash
IMAGE=<seu-registry>/calculator-api:v2      # use uma tag nova a cada deploy
docker build -t $IMAGE .
docker push $IMAGE                           # no kind: kind load docker-image $IMAGE --name <cluster>
kubectl -n calculator set image deployment/calculator-api calculator-api=$IMAGE
kubectl -n calculator rollout status deployment/calculator-api
```

Teste de novo com o comando do passo 2.

**4. Remover**

```bash
kubectl delete -f k8s/
```

Isso apaga o namespace `calculator` e tudo o que há dentro dele.

---

## Aplicação web

| Operação | Endpoint | Exemplo | Resposta |
| --- | --- | --- | --- |
| Adição | `GET /api/sum` | `/api/sum?term_one=4&term_two=1` | `{"result":5}` |
| Subtração | `GET /api/sub` | `/api/sub?term_one=4&term_two=1` | `{"result":3}` |
| Multiplicação | `GET /api/mul` | `/api/mul?term_one=4&term_two=1` | `{"result":4}` |
| Divisão | `GET /api/div` | `/api/div?term_one=4&term_two=1` | `{"result":4}` |
| Healthcheck | `GET /healthz` | `/healthz` | `{"status":"ok"}` |

- Roda na porta `8000`.
- Os termos e o resultado são **inteiros**, como o enunciado define (`<int:result>`). Por isso a divisão é inteira: `5/2 = 2`.
- Em caso de erro, responde `400` com `{"error": "..."}`. Isso vale para termo ausente, termo que não é inteiro e divisão por zero.
- Só é acessível de dentro do cluster: o Service é `ClusterIP`, sem NodePort, LoadBalancer ou Ingress.

Para adicionar uma operação, basta uma linha nova no mapa `operations` do `main.go`. Ela vira o endpoint `/api/<nome>` automaticamente.

### Rodar e testar localmente (sem Kubernetes)

```bash
go run .                                          # sobe em localhost:8000
curl "localhost:8000/api/sum?term_one=4&term_two=1"

go test -cover ./...                              # testes + % de cobertura
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
```

## Deploy

- `k8s/00-namespace.yaml`: o namespace `calculator`.
- `k8s/01-service.yaml`: Service `ClusterIP` na porta 8000.
- `k8s/02-deployment.yaml`: 2 réplicas, probes em `/healthz`, limites de recursos, execução como não-root.

O prefixo numérico garante que o namespace seja criado antes dos outros recursos.

O deploy usa a imagem pública `ghcr.io/raulmoura157/calculator-api:latest`, que o CI publica a cada push na `main`.

A troca de versão não derruba requisições:
- O rollout é gradual e só avança quando o pod novo responde no `/healthz`.
- Ao receber `SIGTERM`, a aplicação continua atendendo por 5 s, tempo para o Kubernetes tirar o pod do Service.
- Depois disso, termina as requisições em andamento antes de sair.

Para chamar a API da sua máquina, abra um túnel:

```bash
kubectl -n calculator port-forward svc/calculator-api 8000:8000
curl "localhost:8000/api/sum?term_one=4&term_two=1"
```

## CI

`.github/workflows/ci.yml` roda `go vet` e os testes e guarda o `coverage.html` como artefato. Na `main`, também publica a imagem no GHCR.
