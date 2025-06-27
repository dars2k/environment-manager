When the task is complete, follow these steps:

1. Run all unit tests.
   - If tests fail, fix the code and re-run until they pass.
   - If tests pass, continue.

2. Build the Docker image.
   - Run: docker compose build, And take only the last 50 lines of output
   - If the output includes an error, fix it and rebuild.
   - If no errors, continue.

3. Start the application.
   - Run: docker compose up -d

4. Verify the application is running as expected.

5. Commit the changes to git.
   - Use a short description starting with one of the following:
     - added
     - removed
     - fix
     - changed
