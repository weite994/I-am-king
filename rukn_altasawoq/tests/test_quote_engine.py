import unittest
from unittest.mock import patch
from rukn_altasawoq.quote_engine.processor import create_quote_from_url

class TestQuoteEngine(unittest.TestCase):

    @patch('rukn_altasawoq.quote_engine.processor.get_product_price_from_url')
    def test_create_quote_from_url(self, mock_get_price):
        # Arrange
        mock_get_price.return_value = 100.0
        customer_id = "test_customer"
        product_url = "http://example.com/product"

        # Act
        quote = create_quote_from_url(customer_id, product_url)

        # Assert
        self.assertEqual(quote.customer_id, customer_id)
        self.assertEqual(quote.product_links, [product_url])
        self.assertEqual(quote.quote_price, 100.0)
        self.assertEqual(quote.markup_percentage, 35.0)
        self.assertEqual(quote.tax, 15.0)
        self.assertEqual(quote.final_price, 150.0)
        self.assertEqual(quote.status, "Pending")

if __name__ == '__main__':
    unittest.main()
