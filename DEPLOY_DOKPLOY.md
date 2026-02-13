# Deploying Translate Management on Dokploy

This guide outlines the steps to deploy the Translate Management application using Dokploy.

## Prerequisites

- A Dokploy instance set up and running.
- Access to your Git repository containing this project.
- A domain name (optional but recommended for SSL).

## Strategy

We will deploy the application as two separate services in Dokploy (Frontend and Backend) plus a Database (PostgreSQL) and a Cache (Redis).

## Step 1: Push Changes

Make sure you have committed and pushed the latest changes to your repository. This guide assumes the codebase includes the recent configuration updates specifically for Dokploy deployment (CORS configuration and Environment Variables support).

## Step 2: Create a Project in Dokploy

1. Log in to your Dokploy Dashboard.
2. Click **"Create Project"**.
3. Name it `translate-management` (or similar).

## Step 3: Deploy Database (PostgreSQL)

1. In your project, go to the **"Database"** tab.
2. Click **"Create Database"** -> **"PostgreSQL"**.
3. Name: `postgres-db` (or similar).
4. Dokploy will generate credentials (Internal Host, Port, User, Password, Database Name). **Note these down**.

## Step 4: Deploy Cache (Redis)

1. In your project, go to the **"Database"** tab.
2. Click **"Create Database"** -> **"Redis"**.
3. Name: `redis-cache` (or similar).
4. Dokploy will generate the Internal Host and Port. **Note these down**.

## Step 5: Deploy Backend

1. Go to the **"Application"** tab.
2. Click **"Create Application"**.
3. Select your Git Provider and Repository.
4. **Configuration**:
   - **Name**: `tm-backend`
   - **Branch**: `main`
   - **Build Type**: `Dockerfile`
   - **Context Path**: `/backend`
   - **Dockerfile Path**: `Dockerfile`
5. **Environment Variables**:
   Add the following variables using the credentials from Steps 3 & 4:

   | Key            | Value                                | Description                                                                  |
   | -------------- | ------------------------------------ | ---------------------------------------------------------------------------- |
   | `DB_HOST`      | `postgres-db` (or internal host)     | Hostname of Postgres service                                                 |
   | `DB_PORT`      | `5432`                               | Postgres Port                                                                |
   | `DB_USER`      | `postgres`                           | Postgres User                                                                |
   | `DB_PASSWORD`  | `<your-password>`                    | Postgres Password                                                            |
   | `DB_NAME`      | `postgres`                           | Database Name (default is usually postgres)                                  |
   | `REDIS_HOST`   | `redis-cache` (or internal host)     | Hostname of Redis service                                                    |
   | `REDIS_PORT`   | `6379`                               | Redis Port                                                                   |
   | `PORT`         | `3000`                               | Application Port                                                             |
   | `FRONTEND_URL` | `https://tm-frontend.yourdomain.com` | URL of your frontend (set this after frontend is deployed or predict it now) |

6. Click **"Create"** and then **"Deploy"**.
7. Once configured, go to **"Domains"** tab and generate a domain (e.g., `api-tm.yourdomain.com`). This will be your **Backend URL**.

## Step 6: Deploy Frontend

1. Go to the **"Application"** tab.
2. Click **"Create Application"**.
3. Select your Git Provider and Repository.
4. **Configuration**:
   - **Name**: `tm-frontend`
   - **Branch**: `main`
   - **Build Type**: `Dockerfile`
   - **Context Path**: `/frontend`
   - **Dockerfile Path**: `Dockerfile`
5. **Environment Variables**:

   | Key                | Value                           | Description                                        |
   | ------------------ | ------------------------------- | -------------------------------------------------- |
   | `PUBLIC_API_URL`   | `https://api-tm.yourdomain.com` | The **External** URL of your backend (from Step 5) |
   | `INTERNAL_API_URL` | `http://tm-backend:3000`        | The **Internal** URL of your backend service       |
   | `PORT`             | `5173`                          | Application Port                                   |

   _Note: `INTERNAL_API_URL` is used for Server-Side Rendering (SSR) to communicate internally within the Docker network._

6. Click **"Create"** and then **"Deploy"**.
7. Go to **"Domains"** tab and generate a domain (e.g., `tm-frontend.yourdomain.com`). Updates `FRONTEND_URL` in backend if needed.

## Step 7: Initialize Database

The application does not automatically run migrations on startup. You need to run the initial SQL scripts.

1. Locate the `migrations/001_initial.sql` file in your repository.
2. In Dokploy, go to your **PostgreSQL Database** -> **"Backup / Restore"** or execute via a client.
3. Alternatively, use a local client (like DBeaver or `psql`) to connect to your remote database (expose it temporarily or use SSH tunnel) and run the scripts in `migrations/`.

## Troubleshooting

- **CORS Errors**: Ensure `FRONTEND_URL` in Backend Env Vars matches strictly the URL in your browser including `https://` and no trailing slash.
- **Connection Refused (SSR)**: Ensure `INTERNAL_API_URL` points to the correct internal service name and port (usually `http://<service-name>:<port>`).
