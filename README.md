# Rukn AlTasawoq - Backend Ops System

This repository contains the backend system for the Rukn AlTasawoq project. It is a modular, AI-enhanced system designed to integrate customer journeys, affiliate systems, order automation, dynamic pricing, and personalized promotions.

## Project Structure

The project is organized into the following directories:

- `rukn_altasawoq/`: The main Python package for the project.
    - `models/`: Contains the Pydantic data models for the various entities in the system (e.g., `Client`, `Order`, `Quote`).
    - `services/`: Contains modules for interacting with external services like Notion, Google Sheets, and ChatbotBuilder.
    - `customer_engine/`: Logic for managing customer profiles and lifecycle.
    - `affiliate_engine/`: Logic for the multi-level affiliate and commission system.
    - `order_engine/`: Logic for handling quotes and orders.
    - `offer_engine/`: Logic for the smart offer and promotion system.
    - `shipping_engine/`: Logic for shipment tracking and logistics.
    - `automation_engine/`: Logic for workflow automation.
    - `analytics_engine/`: Logic for data analytics and reporting.
    - `content_engine/`: Logic for content and campaign generation.
    - `quote_engine/`: Contains the dynamic quote processor.
- `tests/`: Contains unit tests for the project.
- `requirements.txt`: A list of the Python dependencies for this project.

## Getting Started

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   ```

2. **Install the dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Set up environment variables:**
   - Create a `.env` file in the root of the project.
   - Add the following environment variables to the `.env` file:
     ```
     NOTION_API_KEY=your_notion_api_key
     GOOGLE_SHEETS_CREDENTIALS_PATH=/path/to/your/google-sheets-credentials.json
     CHATBOT_BUILDER_API_KEY=your_chatbot_builder_api_key
     ```

4. **Run the tests:**
   ```bash
   python -m unittest discover -s rukn_altasawoq/tests
   ```

## Contributing

Please follow the existing coding style and add unit tests for any new features or bug fixes.
