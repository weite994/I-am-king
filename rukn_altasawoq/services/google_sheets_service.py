import os
import gspread
from oauth2client.service_account import ServiceAccountCredentials

class GoogleSheetsService:
    def __init__(self):
        scope = ["https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"]
        creds = ServiceAccountCredentials.from_json_keyfile_name(os.environ.get("GOOGLE_SHEETS_CREDENTIALS_PATH"), scope)
        self.client = gspread.authorize(creds)

    def get_sheet(self, sheet_name: str):
        return self.client.open(sheet_name).sheet1

    def add_row(self, sheet_name: str, data: list):
        sheet = self.get_sheet(sheet_name)
        sheet.append_row(data)

    def get_all_records(self, sheet_name: str):
        sheet = self.get_sheet(sheet_name)
        return sheet.get_all_records()

    def update_cell(self, sheet_name: str, row: int, col: int, value: str):
        sheet = self.get_sheet(sheet_name)
        sheet.update_cell(row, col, value)
