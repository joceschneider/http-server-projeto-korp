# Projeto Korp — Desafio DevOps

Projeto desenvolvido como solução para um desafio técnico de DevOps, com foco em **Golang, Docker, Docker Compose, NGINX, Prometheus, Grafana e Ansible**.

O objetivo é disponibilizar uma aplicação HTTP simples, executá-la em containers, adicionar monitoramento e observabilidade e, por fim, automatizar todo o provisionamento do ambiente Linux com Ansible.

---

## Visão geral

A solução final possui os seguintes componentes:

- **Aplicação Golang**: expõe o endpoint principal do projeto e métricas.
- **NGINX**: atua como proxy reverso e único ponto de entrada HTTP.
- **Prometheus**: coleta as métricas da aplicação.
- **Grafana**: exibe um dashboard de monitoramento.
- **Ansible**: automatiza a preparação e o provisionamento do ambiente Linux.
- **Docker Compose**: orquestra os containers da aplicação.

---

## Arquitetura

```text
                         Cliente
                            |
                            | HTTP :80
                            v
                      +-----------+
                      |   NGINX   |
                      |    :80    |
                      +-----+-----+
                            |
                            | Docker bridge
                            v
                +--------------------------+
                | http-server-projeto-korp |
                |          :8080           |
                +------------+-------------+
                             |
                   +---------+---------+
                   |                   |
             GET /health          GET /metrics
                   |                   |
                   |                   v
                   |             +------------+
                   |             | Prometheus |
                   |             |   :9090    |
                   |             +------+-----+
                   |                    |
                   |                    v
                   |              +-----------+
                   +------------->|  Grafana  |
                                  |   :3000   |
                                  +-----------+
```

Todos os containers utilizam a rede Docker:

```text
projeto-korp-network
```

---

## Tecnologias utilizadas

- Golang
- Docker
- Docker Compose
- NGINX
- Prometheus
- Grafana
- Ansible
- Ubuntu Linux
- SSH

---

## Estrutura do projeto

```text
http-server-projeto-korp/
├── ansible/
│   ├── inventory.ini
│   └── playbook.yml
├── grafana/
│   ├── dashboards/
│   │   └── http-server-projeto-korp-dashboard.json
│   └── provisioning/
│       ├── dashboards/
│       │   └── dashboards.yml
│       └── datasources/
│           └── datasources.yml
├── nginx/
│   └── conf.d/
│       └── http-server-projeto-korp.conf
├── prometheus/
│   └── prometheus.yml
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── main.go
```

---

# Parte 1 — Serviço e Arquitetura do Ambiente

## Serviço HTTP

A aplicação foi desenvolvida em Golang e executa na porta interna `8080`.

Endpoint principal:

```http
GET /projeto-korp
```

Exemplo de resposta:

```json
{
  "nome": "Projeto Korp",
  "horario": "2026-09-09T15:32:10Z"
}
```

O horário é calculado dinamicamente em UTC a cada requisição.

## Dockerfile

Foi utilizado **multi-stage build** para separar a etapa de compilação da imagem final.

Principais objetivos:

- reduzir o tamanho da imagem;
- não levar o toolchain completo do Go para a imagem final;
- executar a aplicação com usuário não-root;
- manter o processo de build reproduzível.

Build manual:

```bash
docker build -t http-server-projeto-korp:1.0 .
```

## Rede Docker

Foi utilizada uma rede Docker dedicada em modo `bridge`:

```bash
docker network create --driver bridge projeto-korp-network
```

O container da aplicação não publica a porta `8080` diretamente no host.

O NGINX é o único ponto de entrada público da aplicação:

```text
host :80 -> nginx :80 -> http-server-projeto-korp :8080
```

## NGINX

O NGINX atua como proxy reverso.

Arquivo:

```text
nginx/conf.d/http-server-projeto-korp.conf
```

Trecho principal:

```nginx
server {
    listen 80;

    location / {
        proxy_pass http://http-server-projeto-korp:8080;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Teste:

```bash
curl http://localhost:80/projeto-korp
```

---

# Parte 2 — Monitoramento e Observabilidade

A aplicação também expõe:

```text
GET /health
GET /metrics
```

## Health check

```bash
curl http://localhost:80/health
```

Resposta:

```json
{
  "status": "UP"
}
```

## Métricas Prometheus

A aplicação expõe métricas no padrão Prometheus em `/metrics`.

### Volume de requisições

Métrica personalizada:

```promql
projeto_korp_requests_total
```

Ela é incrementada a cada chamada ao endpoint `/projeto-korp`.

### Disponibilidade

A disponibilidade é acompanhada pela métrica nativa do Prometheus:

```promql
up{job="http-server-projeto-korp"}
```

Interpretação:

```text
1 = serviço disponível
0 = serviço indisponível para coleta
```

## Prometheus

Configuração:

```text
prometheus/prometheus.yml
```

O Prometheus coleta as métricas da aplicação a cada 5 segundos.

```yaml
global:
  scrape_interval: 5s

scrape_configs:
  - job_name: "http-server-projeto-korp"
    static_configs:
      - targets:
          - "http-server-projeto-korp:8080"
```

Interface:

```text
http://localhost:9090
```

## Grafana

Interface:

```text
http://localhost:3000
```

Dashboard criado:

```text
Projeto Korp - Monitoramento
```

Painéis:

### Disponibilidade do Serviço

```promql
up{job="http-server-projeto-korp"}
```

Mapeamento visual:

```text
1 -> UP
0 -> DOWN
```

### Total de Requisições

```promql
projeto_korp_requests_total
```

### Taxa de Requisições

```promql
rate(projeto_korp_requests_total[1m])
```

Configuração padrão:

```text
Período: últimos 15 minutos
Refresh: 5 segundos
```

## Provisionamento automático do Grafana

O datasource Prometheus e o dashboard são provisionados automaticamente por arquivos:

```text
grafana/provisioning/datasources/datasources.yml
grafana/provisioning/dashboards/dashboards.yml
grafana/dashboards/http-server-projeto-korp-dashboard.json
```

Dessa forma, em um ambiente novo, o Grafana inicia já com o datasource e o dashboard disponíveis.

## Teste de tráfego

```bash
for i in {1..20}; do
  curl -s http://localhost:80/projeto-korp > /dev/null
  sleep 0.2
done
```

Após o teste:

- o contador total aumenta;
- o gráfico de taxa registra atividade;
- a disponibilidade permanece `UP`.

---

# Parte 3 — Automação com Ansible

O Mac foi utilizado como **control node** e uma VM Ubuntu como **managed node**.

Fluxo:

```text
Mac
 |
 | Ansible + SSH
 v
Ubuntu
 |
 +-- instala/configura Docker
 +-- cria diretório do projeto
 +-- copia arquivos
 +-- garante rede Docker
 +-- executa Docker Compose
 +-- sobe NGINX
 +-- sobe aplicação Go
 +-- sobe Prometheus
 +-- sobe Grafana
 +-- valida /projeto-korp
 +-- valida /health
 +-- exibe resposta no console
```

## Inventory

Exemplo:

```ini
[korp]
linux_korp ansible_host=<IP_DA_VM> ansible_user=<USUARIO> ansible_ssh_private_key_file=<CAMINHO_DA_CHAVE> ansible_python_interpreter=/usr/bin/python3
```

> Ajuste IP, usuário e caminho da chave SSH conforme o ambiente local.
> Não versionar chaves privadas no repositório.

## Testar comunicação com Ansible

```bash
ansible -i ansible/inventory.ini korp -m ping
```

Resultado esperado:

```text
linux_korp | SUCCESS => {
    "changed": false,
    "ping": "pong"
}
```

## Provisionamento completo

```bash
ansible-playbook   -i ansible/inventory.ini   ansible/playbook.yml   --ask-become-pass
```

O playbook:

1. identifica sistema operacional e arquitetura;
2. instala pré-requisitos;
3. configura o repositório oficial do Docker;
4. instala Docker Engine e Docker Compose;
5. garante que o Docker esteja ativo;
6. adiciona o usuário ao grupo Docker;
7. desabilita NGINX instalado diretamente no host, caso exista;
8. cria `/opt/http-server-projeto-korp`;
9. copia os arquivos do projeto;
10. garante a existência da rede Docker;
11. executa build e Docker Compose;
12. aguarda a aplicação responder;
13. valida `/projeto-korp`;
14. valida `/health`;
15. exibe a resposta HTTP no console;
16. exibe os containers em execução.

## Resultado esperado

```text
TASK [Exibir resposta do serviço no console]

{
  "nome": "Projeto Korp",
  "horario": "2026-09-09T15:32:10Z"
}
```

Health:

```text
{
  "status": "UP"
}
```

Resumo:

```text
unreachable=0
failed=0
```

---

# Como executar manualmente com Docker Compose

## Pré-requisitos

- Docker Engine ou Docker Desktop
- Docker Compose
- rede Docker `projeto-korp-network`

Criar a rede, caso ainda não exista:

```bash
docker network create --driver bridge projeto-korp-network
```

Validar:

```bash
docker compose config
```

Subir:

```bash
docker compose up --build -d
```

Verificar:

```bash
docker compose ps
```

Aplicação:

```bash
curl http://localhost:80/projeto-korp
```

Health:

```bash
curl http://localhost:80/health
```

Prometheus:

```text
http://localhost:9090
```

Grafana:

```text
http://localhost:3000
```

Parar:

```bash
docker compose down
```

---

# Decisões técnicas

## Multi-stage build

O build da aplicação Go é realizado em uma imagem de compilação e somente o binário final é copiado para a imagem de runtime.

Benefícios:

- imagem final menor;
- menor superfície de ataque;
- separação clara entre build e runtime.

## Aplicação sem porta publicada

O serviço Go não publica a porta `8080` diretamente no host.

Somente o NGINX publica a porta `80`, mantendo o backend acessível apenas pela rede Docker.

## Rede dedicada

Foi criada uma rede bridge dedicada para comunicação entre:

- NGINX;
- aplicação;
- Prometheus;
- Grafana.

## Métricas

Foi utilizada uma combinação de:

- métrica personalizada para volume de requisições;
- métrica `up` do Prometheus para disponibilidade;
- endpoint `/health` como validação adicional.

## Grafana provisionado por arquivos

O dashboard e o datasource são tratados como configuração versionável, permitindo reproduzir a mesma visualização em ambientes novos.

## Ansible idempotente

As tarefas foram construídas para declarar o estado desejado e evitar recriação desnecessária de recursos já existentes.

---

# Endpoints e portas

| Componente | Endpoint / Porta |
|---|---|
| Aplicação | `GET /projeto-korp` |
| Health | `GET /health` |
| Métricas | `GET /metrics` |
| NGINX | `80` |
| Aplicação Go | `8080` interno |
| Prometheus | `9090` |
| Grafana | `3000` |

---

# Status do projeto

```text
Parte 1 — concluída
Parte 2 — concluída
Parte 3 — concluída
```

O ambiente foi validado manualmente e também através do playbook Ansible.

---

## Observações de segurança

- Não versionar chaves SSH privadas.
- Não armazenar senhas no repositório.
- Ajustar informações específicas do ambiente local no inventory.
- O usuário incluído no grupo `docker` possui privilégios elevados no host Linux.

---

## Próximos passos possíveis

- adicionar healthchecks no Docker Compose;
- adicionar regras de alerta no Prometheus/Grafana;
- incluir CI/CD para validação do projeto;
- versionar imagens Docker com tags imutáveis;
- adicionar testes automatizados da aplicação;
- utilizar Ansible Vault para segredos, caso necessário.
