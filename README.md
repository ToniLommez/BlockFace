# BlockFace

O projeto BlockFace é um sistema de detecção e monitoramento de rostos que entram e saem de estabelecimentos, com armazenamento seguro dos dados em uma blockchain. O sistema combina inteligência artificial avançada e tecnologia de blockchain para garantir a imutabilidade e a segurança das informações coletadas. A blockchain implementa um modelo híbrido de consenso, utilizando prova de trabalho para validação inicial e prova de autoridade para inserção de novos blocos. Os líderes da rede são responsáveis por realizar as inserções, e cada nova inserção estende seu período de liderança, promovendo estabilidade e eficiência no processo.

Para o reconhecimento facial, o sistema emprega dois modelos de IA. O YOLO é utilizado para detectar rostos nas imagens, enquanto o FaceNet, modelo pré-treinado do Google, realiza o reconhecimento e a identificação dos indivíduos. Foi realizado um fine-tuning do FaceNet com a base de dados QMUL-SurvFace, permitindo um desempenho otimizado para o contexto do projeto.

## Alunos integrantes da equipe

* Bernardo Marques Fernandes
* Eric Miranda Ferreita Guimarães
* Marcos Antônio Lommez Cândido Ribeiro
* Saulo de Moura Zandona Freitas

## Professores responsáveis

* Alexei Manso Correa Machado
* Fatima de Lima Procopio Duarte
* Henrique Cota de Freitas

## Instruções de utilização

para baixar as bibliotecas necessárias: pip install -r Codigo/requirements

baixe os modelos treinados para configuração em: https://drive.google.com/file/d/1px2trydOx1tUigFCiDUaZJZf1i6tFcCZ/view?usp=sharing
adicione os arquivos em Codigo/data/facenet_config

### blockChain

#### Inicialização
go run main.go
obs.: precisa estar na pasta Codigo


#### Execução de terminal(algoritmo rodando)
# Comandos Disponíveis na Nether Blockchain

Abaixo estão listados os comandos disponíveis para interagir com o sistema da Nether Blockchain:

| Comando              | Descrição                                                                       |
|----------------------|---------------------------------------------------------------------------------|
| `help`               | Printa as funções disponíveis no console.                                       |
| `clear`              | Limpa o conteúdo do terminal.                                                   |
| `register`           | Registra uma nova conta na Nether Blockchain.                                   |
| `login`              | Login de uma conta existente na Nether Blockchain.                              |
| `test userdata`      | Testa a sanidade dos dados de registro.                                         |
| `see userdata`       | Visualiza os dados do usuário atual.                                            |
| `new blockchain`     | Cria uma nova Blockchain.                                                       |
| `load blockchain`    | Carrega uma blockchain da memória secundária para a memória primária.           |
| `show blockchain`    | Printa a blockchain no terminal.                                                |
| `start server`       | Inicializa um servidor para conexão peer-to-peer (p2p).                         |
| `start client`       | Inicializa um cliente para conexão peer-to-peer (p2p).                          |
| `ping all`           | Envia um PING broadcast e recebe um PONG para mostrar uma conexão estabelecida. |
| `start election`     | Um líder atual inicia a eleição para determinar novos líderes.                  |
| `show connections`   | Mostra todos os nós conectados ao sistema.                                      |
| `download blockchain`| Faz o download da versão da blockchain contida nos líderes.                     |
| `start endpoint`     | Abre conexão para recebimento de informações dos serviços de câmera.            |
| `exit`               | Sai da Nether Blockchain.                                                       |

### Para inicializar o algoritmo e a aplicação para as câmeras de segurança: 
- streamlit run main.py
  obs.: precisa estar na pasta Codigo
- Na aplicação, escolha a câmera a esquerda para iniciar a visualização
