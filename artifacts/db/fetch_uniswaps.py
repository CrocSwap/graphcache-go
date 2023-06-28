import requests
import json
import time

import sqlite3

conn = sqlite3.connect("mydatabase.db")  # Replace with the name of your database file
cursor = conn.cursor()
# Format of the swaps table
#     CREATE TABLE swaps (
#   id INTEGER PRIMARY KEY,
#   swap JSON,
#   swap_time DATETIME,
#   swap_id STRING
# );

# define the endpoint
url = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"

# Start from the current time and keep fetching swaps that are older
timestamp = int(time.time())  # Current timestamp

# The timestamp you want to query back to (e.g., 1672444800 is for 1st January 2023)
end_timestamp = 1687279382


total_swaps = 0
# Fetch the earliest swap in the database
cursor.execute(
    """
SELECT swap_time
FROM swaps
ORDER BY swap_time ASC
LIMIT 1;
"""
)

earliest_swap = cursor.fetchall()

if len(earliest_swap) > 0:
    timestamp = earliest_swap[0][0]
    print(
        f"Starting from {time.strftime('%m/%d/%Y %H:%M:%S', time.localtime(timestamp))}"
    )


while timestamp > end_timestamp:
    query = f"""
    {{
      swaps(where: {{timestamp_lte: {timestamp}, pool_in: ["0x7bea39867e4169dbe237d55c8242a8f2fcdcc387", "0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", "0xe0554a476a092703abdb3ef35c80e0d76d32939f"]}}, first: 1000, orderBy: timestamp, orderDirection: desc) {{
            id
            transaction {{
                id
                blockNumber
            }}
            pool {{
                id
                token0 {{
                    id
                    symbol
                }}
                token1 {{
                    id
                    symbol
                }}
            }}
            amount0
            amount1
            timestamp
      }}
    }}
    """


    response = requests.post(url, json={"query": query})

    print(
        f"Querying swaps before {time.strftime('%m/%d/%Y %H:%M:%S', time.localtime(timestamp))}..."
    )

    # check the status of the request
    if response.status_code == 200:
        # convert the response to JSON
        data = json.loads(response.text)
        swaps = data["data"]["swaps"]
        print(f"Got {len(swaps)} swaps")

        for swap in swaps:
            cursor.execute(
                "INSERT INTO swaps (swap, swap_time, swap_id) VALUES (?, ?, ?)",
                (json.dumps(swap), swap["timestamp"], swap["id"]),
            )
        conn.commit()

        # if there are no more swaps, break the loop
        if len(swaps) == 0:
            break

        # update the timestamp to be the timestamp of the last swap in the result
        timestamp = int(swaps[-1]["timestamp"])

    else:
        print(f"Query failed with status code {response.status_code}")
        break


