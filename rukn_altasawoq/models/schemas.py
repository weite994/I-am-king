from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime

from uuid import uuid4

class Client(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    name: str
    phone: str
    sessions: List[str] = []
    addresses: List[str] = []
    wallet_id: Optional[str] = None
    affiliate_level: str = "Free"
    loyalty_points: int = 0
    preferences: dict = {}
    behavior_history: List[str] = []
    customer_type: str = "New"
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Quote(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    customer_id: str
    product_links: List[str] = []
    images: List[str] = []
    cart: List[dict] = []
    voice_note: Optional[str] = None
    quote_price: float
    markup_percentage: float = 35.0
    tax: float = 0.0
    discount: float = 0.0
    final_price: float
    urgency_tag: Optional[str] = None
    status: str = "Pending"  # Pending, Approved, Rejected
    handler_id: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Order(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    quote_id: str
    customer_id: str
    products: List[dict]
    total_price: float
    payment_status: str = "Unpaid"  # Unpaid, Paid, Refunded
    order_status: str = "Confirmed"  # Confirmed, Shipped, Delivered, Canceled
    shipping_id: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Wallet(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    customer_id: str
    balance: float = 0.0
    transactions: List[dict] = []
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Coupon(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    code: str
    discount_percentage: Optional[float] = None
    discount_amount: Optional[float] = None
    valid_from: datetime
    valid_to: datetime
    is_active: bool = True
    usage_limit: int = 1
    used_count: int = 0
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Product(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    name: str
    description: Optional[str] = None
    price: float
    currency: str = "USD"
    supplier: Optional[str] = None
    product_url: Optional[str] = None
    images: List[str] = []
    attributes: dict = {}
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Shipment(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    order_id: str
    tracking_number: str
    provider: str
    status: str = "Pending"  # Pending, In-Transit, Delivered, Failed
    tracking_history: List[dict] = []
    media: dict = {
        "invoice": Optional[str],
        "proof_of_delivery": Optional[str],
        "waybill": Optional[str],
        "warehouse_receipt": Optional[str],
    }
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class Affiliate(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid4()), alias="_id")
    client_id: str
    level: str  # Free, Paid, Sponsored
    upline_id: Optional[str] = None
    downline_ids: List[str] = []
    commission_rate: float
    commission_balance: float = 0.0
    performance_score: float = 0.0
    linked_offers: List[str] = []
    traffic_sources: List[str] = []
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
