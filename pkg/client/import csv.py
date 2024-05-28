import csv
import webbrowser
import os

def open_links_in_browser(csv_file):
    with open(csv_file, 'r') as file:
        reader = csv.DictReader(file)
        link_count = 0
        for row in reader:
            for key, value in row.items():
                if key.startswith('link'):
                    link = value.strip()
                    if link:
                        webbrowser.open(link)
                        link_count += 1
                        if link_count == 20:
                            return  

if __name__ == "__main__":
    csv_file = os.path.join(os.path.dirname(__file__), 'job_openings.csv')
    open_links_in_browser(csv_file)