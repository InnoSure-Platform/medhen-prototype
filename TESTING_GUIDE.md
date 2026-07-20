# Medhen Platform: Real-Time Testing Guide

This guide provides a structured, step-by-step walkthrough to boot up the entire Medhen Platform (infrastructure, backend services, API gateway, and frontend) and verify the integration in real-time.

---

## Step 1: Boot Up the Backend & Infrastructure

The platform is a single **modular-monolith** service (`medhen-api`). Start the infra backbone
(Keycloak, PostgreSQL, Kafka, Valkey) via Docker, then run the monolith.

1. Open a new terminal at the root of the project: `~/Documents/ITG/Medhen Prototype/`
2. Start infrastructure and the API:
   ```bash
   make infra-up
   export DATABASE_URL="postgres://medhen:medhen@localhost:5432/medhen?sslmode=disable"
   export TELEBIRR_WEBHOOK_SECRET="dev-secret"
   make api            # builds + runs medhen-api on :8080
   ```
3. **Verification**:
   - Run `docker ps` to ensure `medhen-keycloak-1`, `medhen-postgres-1`, etc., are running.
   - The API logs `medhen-api listening` with the list of registered modules.
   - Ping the health endpoint:
     ```bash
     curl -s http://localhost:8080/healthz
     ```

## Step 2: Start the Next.js Frontend

With the backend active, start the `pc-web` frontend in a separate terminal.

1. Open a **second terminal** and navigate to the web directory:
   ```bash
   cd "~/Documents/ITG/Medhen Prototype/web"
   ```
2. Start the development server (make sure port 3000 is free):
   ```bash
   npm run dev
   ```
3. **Verification**:
   - The terminal should state `Ready in Xms` and be listening on `http://localhost:3000`.

## Step 3: Test Keycloak Single Sign-On (SSO)

Now we will test the enterprise authentication flow.

1. Open your web browser and navigate to **[http://localhost:3000](http://localhost:3000)**.
2. Click **Sign In**. 
   - You should be instantly redirected to the Keycloak server at `http://localhost:8081/realms/medhen/protocol/openid-connect/auth...`
3. Enter the test broker credentials provided in the realm config:
   - **Username:** `demo-agent`
   - **Password:** `medhen-demo`
4. Click **Log In**.
5. **Verification**:
   - You should be redirected back to Next.js (`http://localhost:3000`).
   - The NextAuth middleware will extract the `agent` role from the Keycloak JWT.
   - If you attempt to navigate to a protected route (e.g., `http://localhost:3000/broker`), you will be granted access. 
   - If you try to navigate to `http://localhost:3000/admin`, the middleware will block you and redirect you back to `/login` because your role is strictly `agent`/`broker`.

## Step 4: Test Different Roles

To verify Role-Based Access Control (RBAC), log out and try another user.

1. **Log out** of the `demo-agent` account.
2. **Log in** with the Claims Handler credentials:
   - **Username:** `demo-claims`
   - **Password:** `medhen-demo`
3. **Verification**: The system should now identify you as a `claims` user, giving you access to claims-specific routes while blocking you from broker-specific routes.

## Troubleshooting

- **CSS Missing / Unstyled UI?** Ensure you don't have an old Next.js process running on port 3001. Terminate all terminal windows running Next.js and run `npm run dev` cleanly.
- **Keycloak Login Fails?** Ensure the `KEYCLOAK_SECRET` matches between your `web/.env.local` and the `realm-medhen.json` file, and that the Keycloak container successfully restarted.
- **Port Conflicts?** Run `lsof -t -i:3000 | xargs kill -9` to forcefully free up port 3000 if Next.js fails to bind.
