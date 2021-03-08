Quick boilerplate code for building APIs in Go with gofiber, redis, sqlboiler, postgres and dbmate.

> Features -

1. User Registration And Give Role
2. User Email Verification & Account Activation
3. User Password Reset
4. OTP Generation And Send Via Email
5. User Login And Generate Tokens (Paseto)
6. Role Base Access Control
7. Refresh Tokens
8. Logout And Delete Tokens

> Setup -

1. Edit _sqlboiler.toml_ and _.env_ file and enter database details.
2. Make other changes as you want.
3. Then run
    - **_make migration-up_** - it will create database and run migrations.
    - **_make boil_** - to generate type safe sqlboiler code.
    - **_make redis_** - to start redis server.
    - **_make server_** - to run development server.
