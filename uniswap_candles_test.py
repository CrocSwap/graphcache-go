import requests
import unittest
import datetime

# Generate timestamps for the last 4 weeks (once per week)
weekly_cases = [(datetime.datetime.now() - datetime.timedelta(weeks=i)).timestamp() for i in range(1, 5)]

# Generate timestamps for the last 6 months (once per month)
monthly_cases = [(datetime.datetime.now() - datetime.timedelta(weeks=4*i)).timestamp() for i in range(2, 7)]

# Combine the two lists into the test_cases list
test_cases = [{"n": 100, "time": int(time)} for time in weekly_cases + monthly_cases]


BASE = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
QUOTE = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
POOLIDX = 36000
PERIOD = 14400
CHAINID = "0x1"

LOCAL_URL = "http://localhost:8080"
STAGING_URL = "http://34.173.105.247:8080"



class TestAPI(unittest.TestCase):
    def test_endpoint(self):
        # Define test cases for n and time values
        for i,case in  enumerate(test_cases):
            n = case["n"]
            time = case["time"]
            # Print new line
            print()
            print(f"test #{i}: {datetime.datetime.fromtimestamp(time)}")

            url = f"{STAGING_URL}/gcgo/pool_candles?base={BASE}&quote={QUOTE}&poolIdx={POOLIDX}&period={PERIOD}&n={n}&time={time}&chainId={CHAINID}"
            print(f"url: {url}")
            response = requests.get(url)
            data = response.json()

            if(len(data["data"]) == 0):
                print(f"n: {n}, res: {len(data['data'])}, time: {time}")
                continue

            # print the n assertion
            print(f"n: {n}, res: {len(data['data'])}")

            # self.assertEqual(len(data["data"]), n, "Length of data array does not match n")


            first_obj_time = data["data"][0]["time"]
            print(f"first_obj_time: {first_obj_time}, time: {time}, delta: {first_obj_time - time}")
            # self.assertEqual(first_obj_time, time, "Time of the first object does not match the time of the test case")

            last_obj_time = data["data"][-1]["time"]
            expected_last_obj_time = time - n * PERIOD
            print(f"last_obj_time: {last_obj_time}, expected_last_obj_time: {expected_last_obj_time}, delta: {last_obj_time - expected_last_obj_time}")
            # self.assertEqual(last_obj_time, expected_last_obj_time, "Time of the last object does not match expected time")

if __name__ == "__main__":
    unittest.main()
