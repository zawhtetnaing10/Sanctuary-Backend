# Sanctuary-Backend

This project serves as the backend for the Sanctuary chat application, providing robust user management, authentication, conversation handling, real-time messaging via WebSockets, and file storage capabilities.

## Features

* **User Management:** Create, retrieve, and manage user profiles.
* **Authentication & Authorization:** Secure user login using JWT (JSON Web Tokens) with Bearer token authentication.
* **Post Management:** Create, browse, like, and comment on user posts.
* **Friend Requests:** Send and accept friend requests, and browse the friends list.
* **Conversation Management:** Create and retrieve chat conversations, including group chats and direct messages.
* **Real-time Chat:** Instant message delivery via WebSockets. Messages are persisted to the database and broadcasted to relevant participants.
* **File Uploads:** Securely upload and retrieve files (e.g., images, documents) using AWS S3.

## Technologies Used

* **Go:** The primary programming language for the backend.
* **PostgreSQL:** Relational database for data storage.
* **`sqlc`:** Generates type-safe Go code from SQL queries, ensuring robust database interactions.
* **`gorilla/websocket`:** Go package for building WebSocket servers.
* **AWS S3 SDK for Go:** For interacting with Amazon S3 for file storage.
* **JWT:** For secure token-based authentication.
* **`zap`:** Structured logging.

## Setup Instructions

Follow these steps to get the backend running on your local machine.

### Prerequisites

* **Go:** [Install Go](https://golang.org/doc/install) (version 1.20 or higher recommended).
* **PostgreSQL:** [Install PostgreSQL](https://www.postgresql.org/download/) (version 14 or higher recommended).
* **AWS CLI / AWS SDK configured:** Ensure you have an AWS account and your local environment is configured with AWS credentials (e.g., via `~/.aws/credentials` or environment variables) that an IAM user can use.

### Database Setup

1.  **Create PostgreSQL Database:**
    You need to create a new PostgreSQL database named `sanctuary`.
    ```bash
    psql -U your_postgres_user -c "CREATE DATABASE sanctuary;"
    ```
    (Replace `your_postgres_user` with your PostgreSQL username)

2.  **Run Migrations:**
    Navigate to the `sql/schema` directory in your project root. Execute the SQL migration files to set up the necessary tables.
    ```bash
    # Example using psql (assuming your schema files are like 001_init.sql, 002_chat_messages.sql etc.)
    psql -U your_postgres_user -d sanctuary -f sql/schema/001_init.sql
    psql -U your_postgres_user -d sanctuary -f sql/schema/002_chat_messages.sql
    # ... repeat for all your schema files in order based on naming convention or dependencies
    ```
    **Note:** If you are using a migration tool like `goose` or `migrate`, follow its instructions to apply all pending migrations.

### AWS S3 Setup

This project uses AWS S3 for storing files (like user avatars or chat attachments).

1.  **Create an IAM User:**
    * Go to the AWS IAM Console.
    * Create a new IAM user specifically for your application.
    * Generate an **Access Key ID** and a **Secret Access Key** for this user. **Save these securely**, as you will need them for your `.env` file.
    * Attach a policy to this user that grants necessary S3 permissions. At a minimum, it should have permissions for `s3:PutObject`, `s3:GetObject`, and `s3:DeleteObject` on your designated bucket. A common approach is to grant access to a specific bucket:
        ```json
        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": [
                        "s3:PutObject",
                        "s3:GetObject",
                        "s3:DeleteObject"
                    ],
                    "Resource": "arn:aws:s3:::your-sanctuary-bucket-name/*"
                }
            ]
        }
        ```
        (Replace `your-sanctuary-bucket-name` with your actual bucket name)

2.  **Create a Public S3 Bucket:**
    * Go to the AWS S3 Console.
    * Create a new S3 bucket.
    * **Crucially, ensure this bucket is configured for public read access** if you intend to serve images/files directly via S3 URLs (e.g., for user avatars). This usually involves configuring a Bucket Policy. For production, you might serve files through your backend for more control, but for simple setup, public read is common.
    * Note down your **bucket name** and **AWS region** (e.g., `us-east-1`).

### Environment Variables (`.env` file)

Create a file named `.env` in the root directory of your project. You need to replace the placeholder values with your actual data.

```
    DB_URL=
    PLATFORM=DEV
    TOKEN_SECRET=

    AWS_ACCESS_KEY_ID=
    AWS_SECRET_ACCESS_KEY=
    S3_BUCKET=
    S3_REGION=
```

**Explanation of Variables:**

* **`DB_URL`**: Your PostgreSQL connection string. Format: `postgres://user:password@host:port/database_name?sslmode=disable`
    * Example: `postgres://sanctuary_user:mypassword@localhost:5432/sanctuary?sslmode=disable`
* **`PLATFORM`**: Indicates the environment (e.g., `DEV`, `PROD`). Used for conditional logic (e.g., logging levels).
* **`TOKEN_SECRET`**: A long, random, and highly secret string used to sign your JWTs. **Generate a strong, unique one.**
* **`AWS_ACCESS_KEY_ID`**: The Access Key ID of the IAM user you created for S3.
* **`AWS_SECRET_ACCESS_KEY`**: The Secret Access Key of the IAM user you created for S3.
* **`S3_BUCKET`**: The name of the S3 bucket you created (e.g., `sanctuary-chat-files`).
* **`S3_REGION`**: The AWS region where your S3 bucket is located (e.g., `us-east-1`, `ap-southeast-1`).

### Running the Application

1.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```
2.  **Run the Backend:**
    ```bash
    go run .
    ```
    The server should start, typically listening on port `8080` (or as configured).

## API Endpoints

This backend provides a RESTful API for various services, along with a WebSocket endpoint for real-time communication. All API interactions use Bearer Token authentication via JWTs.

For detailed request/response schemas and examples, please refer to the Postman collection.

### Endpoint Categories:

* **Authentication (Auth):**
    * User registration and login.
    * Token generation and validation.
* **User Management:**
    * Retrieving user profiles (e.g., by ID).
    * User image upload and management (via S3 integration).
* **Post Management:**
    * Creating new posts.
    * Browse posts (e.g., user feeds, public feeds).
    * Liking posts.
    * Commenting on posts.
* **Friend Requests:**
    * Sending friend requests to other users.
    * Accepting or rejecting incoming friend requests.
    * Browse the list of established friends.
* **Conversation Management:**
    * Creating new chat conversations (private/group).
    * Retrieving a user's conversations.
    * Adding/removing participants from conversations.
* **Chat Messages:**
    * Retrieving chat message history for a conversation.
    * Sending and receiving real-time messages via WebSockets.
* **WebSocket Endpoint:**
    * **`/ws`**: The primary endpoint for real-time chat. Clients connect using `wss://` (or `ws://` for local development) and provide authentication via Bearer token in the handshake header, along with `conversation_id` in URL query parameters.
