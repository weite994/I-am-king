import requests
from bs4 import BeautifulSoup
from rukn_altasawoq.models.schemas import Quote

def get_product_price_from_url(url: str) -> float:
    """
    Fetches the price of a product from a given URL.
    This is a placeholder and would need to be adapted for specific websites.
    """
    try:
        response = requests.get(url)
        response.raise_for_status()
        soup = BeautifulSoup(response.text, 'html.parser')

        # This is a placeholder for price extraction logic.
        # In a real implementation, you would need specific selectors for each site.
        # For example, for Amazon, it might be:
        # price = soup.find('span', {'id': 'priceblock_ourprice'}).text
        # For now, we'll return a dummy price.
        return 100.0
    except requests.exceptions.RequestException as e:
        print(f"Error fetching product URL: {e}")
        return 0.0

def create_quote_from_url(customer_id: str, product_url: str) -> Quote:
    """
    Creates a quote for a given customer and product URL.
    """
    price = get_product_price_from_url(product_url)

    markup = 0.35  # 35%
    tax = 0.15     # 15%

    quote_price = price
    final_price = price * (1 + markup + tax)

    quote = Quote(
        customer_id=customer_id,
        product_links=[product_url],
        quote_price=quote_price,
        markup_percentage=markup * 100,
        tax=tax * 100,
        final_price=final_price,
        status="Pending"
    )

    return quote
