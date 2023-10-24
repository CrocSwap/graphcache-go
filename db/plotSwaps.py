import sqlite3
import matplotlib.pyplot as plt
import json
import glob
import pandas as pd
import numpy as np


def sort_by_timestamp(timestamps, prices):
    # Zip the timestamps and prices together
    combined = list(zip(timestamps, prices))
    
    # Sort the combined list by the timestamps
    combined.sort(key=lambda x: x[0])
    
    # Unzip the sorted combined list back into separate lists for timestamps and prices
    sorted_timestamps, sorted_prices = zip(*combined)
    
    return sorted_timestamps, sorted_prices

# Path to the shards folder
shards_path = './shards/*.db'

token_ids_to_filter = ['0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48','0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2']

timestamps = []
prices = []

# Iterate over all the database files in the shards folder
for i, db_path in enumerate(glob.glob(shards_path)):

    # if i > 10:
    #     break
    # Connect to the SQLite database
    connection = sqlite3.connect(db_path)
    cursor = connection.cursor()

    # Fetch data from the "swaps" table
    cursor.execute('''
            SELECT *
            FROM swaps
            ''')
    swaps = cursor.fetchall()

    # filter out the data
    filtered_swaps = []
    for data in swaps:
        swap = json.loads(data[1])
        token0_id = swap["pool"]["token0"]["id"]
        token1_id = swap["pool"]["token1"]["id"]

        if token0_id in token_ids_to_filter and token1_id in token_ids_to_filter:
            filtered_swaps.append(swap)

    print(f"Processing {db_path}: {len(filtered_swaps)} out of {len(swaps)} swaps kept")

    for row in filtered_swaps:
        amount0 = float(row["amount0"])
        amount1 = float(row["amount1"])
        base = row["pool"]["token0"]["id"]
        quote = row["pool"]["token1"]["id"]
        base_flow = amount0  # Assuming amount0 is provided elsewhere in the code
        quote_flow = amount1  # Assuming amount1 is provided elsewhere in the code

        base_check = base.lower()

        if base.lower() != "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" and quote.lower() != "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2":
            print(f"Neither base nor quote is weth: {base} {quote}")
        if base == quote: 
            print(f"Base and quote are the same: {base} {quote}")
            continue
        if base.lower() == "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2":
            base_flow, quote_flow = quote_flow, base_flow

        if base_flow == 0 or quote_flow == 0:
            continue

        price = abs(base_flow / quote_flow)

        prices.append(price)
        timestamps.append(row["timestamp"])

    # Close the database connection
    connection.close()

sorted_timestamps, sorted_prices = sort_by_timestamp(timestamps, prices)


# Convert the sorted timestamps and prices to Pandas Series
timestamps_series = pd.Series(sorted_timestamps)
prices_series = pd.Series(sorted_prices)

# Calculate the rolling standard deviation (e.g., over a window of 50 data points)
# rolling_std = prices_series.rolling(window=50).median()
rolling_mad = prices_series.rolling(window=50).apply(lambda x: np.median(np.abs(x - np.median(x))), raw=True)

# Plot the original prices
plt.figure(figsize=(15, 6))
plt.plot(prices_series, label='Price', color='b', linewidth=1)
plt.yscale('log')  # Set y-axis to logarithmic scale
plt.ylabel('Price (Log Scale)')
plt.title('Price Variation and Rolling Standard Deviation over Time')

# Plot the rolling standard deviation
plt.plot(rolling_mad, label='Rolling Std Dev', color='r', linestyle='--', linewidth=1)

plt.legend()
plt.grid(True, linestyle='--', alpha=0.5)
plt.xticks(rotation=45)
plt.tight_layout()

# Display the graph
plt.show()
