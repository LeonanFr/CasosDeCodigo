# 🕵️‍♂️ Casos de Código – API

**Casos de Código** é um *Serious Game* de investigação criminal onde o jogador assume o papel de um **analista forense digital**.  
O diferencial? A principal ferramenta de investigação não é uma lupa, mas o **SQL**.

Esta API orquestra **puzzles narrativos** onde cada pista deve ser extraída, filtrada ou organizada através de consultas a bancos de dados reais, transformando o aprendizado de banco de dados em uma experiência imersiva de *storytelling*.

---

## 🚀 Tecnologias Utilizadas

- **Linguagem:** Go (Golang)  
  Alta performance para gerenciar sessões de jogo e concorrência.

- **Banco de Dados Principal:** MongoDB (Atlas)  
  Persistência de progressão, histórico de comandos e metadados dos casos.

- **Engine de Execução:** SQLite (*in-memory*)  
  Cada sessão gera um banco relacional temporário e isolado para as queries do jogador (*sandbox*).

- **Autenticação:**  
  JWT para usuários registrados e suporte a *Guest Players* via identificadores únicos de sessão (`X-Guest-ID`).

---

## 🧩 Como Funciona?

O motor do jogo processa comandos através de um sistema de regras definido em arquivos JSON:

- **Puzzles Lógicos**  
  Desafios que exigem manipulação de dados através de comandos DML (`UPDATE`, `INSERT`, `DELETE`).

- **Validação em Tempo Real**  
  O sistema verifica o estado do SQLite após cada comando para validar se a solução foi atingida.

- **Foco Narrativo**  
  Sistema de *foco* (`CurrentFocus`) que integra a interação com o cenário (ex: *OLHAR QUADRO*) à lógica do banco de dados.

---

## 📁 O Caso: O Assassinato do DBA

No primeiro caso disponível, o jogador deve investigar a morte de **Marcos**, o DBA chefe de uma agência de inteligência.

- **Investigação Forense**  
  Análise de fios de cabelo (filtros de texto), pegadas (filtros numéricos) e projetos secretos (JOINs complexos).

- **Persistência**  
  Suporte total a jogadores convidados (*Guest*), mantendo o progresso entre sessões via cabeçalho customizado.

---

## 🛠️ Instalação e Configuração

### Variáveis de Ambiente

Para o funcionamento da API, é necessário configurar um ambiente contendo:

- `JWT_SECRET`
- `MONGO_URI` (MongoDB Atlas)
- `MONGO_DB`
- `PORT`

### Execução

O projeto utiliza módulos oficiais do Go.  
Para rodar, basta garantir que as dependências foram baixadas e iniciar a aplicação através do arquivo principal na raiz do diretório.

---

## 🔭 Telemetria Educacional

O projeto permite a coleta de dados para análise pedagógica do aprendizado:

- **Mapeamento de Erros**  
  Identificação de falhas de sintaxe SQL recorrentes.

- **Curva de Aprendizado**  
  Tempo médio de resolução por puzzle e volume de tentativas.

- **Engajamento**  
  Análise de retenção de jogadores convidados vs. registrados.

---

## 👤 Desenvolvedores

**Leonan Freitas**  
https://github.com/LeonanFr

Estudante de Engenharia de Software na UEPA, desenvolvedor de *Serious Games*.

**Leonardo Victor**
https://github.com/Leovical
