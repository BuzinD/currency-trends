# Currency-Trends-Monitor

**What's this?**  
This is a Go project designed to explore and test the capabilities of Go Lang, particularly in the realm of cryptocurrency data monitoring and processing.

---

## **Overview**
Currency-Trends-Monitor is a service that:
- Fetches and processes cryptocurrency data from OKX.
- Monitors real-time trade information using WebSockets.
- Streams real-time trade data to a **Kafka cluster** for further processing.
- Utilizes **Goroutines** for multitasking and concurrent data processing.

---

## **Technologies Used**
- **Go (Golang)** — For building a high-performance and concurrent application.
- **Kafka Cluster** — For reliable data streaming and processing with two broker nodes.
- **Zookeeper** — For managing Kafka brokers and leader election.
- **Docker & Docker Compose** — To simplify local development and deployment.
- **PostgreSQL** — As the main database for storing historical and real-time trade data.
- **WebSocket** — To receive real-time trade updates.
- **Makefile** — For easy task automation and project management.

---

## **New Features**
### 1. **Kafka Integration**
- **Kafka Cluster with 2 Nodes:** The service uses a Kafka cluster with two brokers to ensure high availability and fault tolerance.
- **Asynchronous Data Streaming:** Utilizes the [IBM Sarama](https://github.com/IBM/sarama) library for asynchronous data streaming to Kafka.
- **Real-Time Trade Data:** Streams live trade data to a Kafka topic (`trades`) for further processing or real-time analytics.

### 2. **Multitasking with Goroutines**
- **Concurrent Data Fetching:** Goroutines are leveraged for non-blocking, concurrent fetching of trade data and market updates.
- **Real-Time WebSocket Handling:** Efficiently manages multiple WebSocket connections simultaneously to receive trade updates in real-time.
- **Asynchronous Kafka Producers:** Uses Goroutines to handle asynchronous message production to Kafka, improving throughput and reliability.

### 3. **Improved Architecture**
- **Modular Structure:** Organized with separate packages for requests, responses, services, and configurations.
- **Scalable Design:** Easily extendable to support more cryptocurrencies or additional data processing components.

---

## **Getting Started**

### **1. Clone the Project**
```sh
git clone https://github.com/BuzinD/currency-trends.git
cd currency-trends
```

### **2. Prepare Environment Variables**
Before the first run, update the environment variables in:
```
data-fetcher/env/okx.env
```

### **3. Start the Service for Development**
```sh
make first-run-dev
```
This command:
- Builds Docker images for the database, migrations, Kafka, and the app.
- Runs containers for:
  - **PostgreSQL** (for data storage)
  - **Kafka Cluster** (with Zookeeper and two broker nodes)
  - **App** (Golang service)
  - **Migrations** (for database schema setup)

---

## **Kafka Setup and Configuration**
### **Cluster Configuration**
- The Kafka cluster consists of **2 brokers** (`kafka1` and `kafka2`) managed by Zookeeper.
- Brokers are configured for **high availability** and **fault tolerance**.

### **Viewing Kafka Messages**
To consume messages from the `trades` topic:
```sh
docker-compose exec kafka1 kafka-console-consumer \
  --bootstrap-server kafka1:9092 \
  --topic trades \
  --from-beginning
```

### **Creating and Managing Topics**
```sh
docker-compose exec kafka1 kafka-topics \
  --create \
  --topic trades \
  --partitions 1 \
  --replication-factor 2 \
  --bootstrap-server kafka1:9092
```

### **Adding Kafka brokers to your hosts**
Open `/etc/hosts` file and add the following lines:
```
127.0.0.1 kafka1
127.0.0.1 kafka2
```
---

## **Available Commands**
To view all available commands:
```sh
make help
```

### **Common Commands**
- **Start the application**: `make start`
- **Stop the application**: `make down`

---

## **Features**
### **Data Fetcher**
- **Available Currencies Fetching** — Scheduled **twice a day**.
- **Trade History Candles** — Scheduled **every hour**.
- **Real-Time Trade Data** — Using WebSocket, the service:
  - Receives real-time trade data for configured cryptocurrency pairs.
  - Streams this data asynchronously to **Kafka** for processing or analytics.

### **Concurrency and Performance**
- **Goroutines** are used for:
  - Concurrent data fetching and processing.
  - Asynchronous message production to Kafka.
  - Efficient handling of WebSocket connections.

### **Scalable and Modular Design**
- The project is organized with:
  - `okx/` package for OKX-specific services.
  - `okx/request` and `okx/response` for request/response models.
  - `kafka/` package for Kafka producers and consumers.

---

## **License**
This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.

---

## **Contact**
For questions or collaboration, reach out to **[dmitrybuzin]** at **[gmail]** **[dot]** **[com]** or open an issue on the repository.

