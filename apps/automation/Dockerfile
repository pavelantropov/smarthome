# Build stage
FROM python:3.9-slim AS builder

# Set working directory
WORKDIR /app

# Copy app files
COPY app.py .

# Download dependencies
RUN pip install flask flask-cors

# Expose the port the app runs on
EXPOSE 8081

# Command to run the application
CMD ["python", "app.py"]