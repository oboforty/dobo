import os
from math import log
from datetime import datetime

# set this to number of JWTs issued per day
REVOKE_MAX_TOKENS = os.environ.get('REVOKE_MAX_TOKENS', 5000000)
REVOKE_P_COLL = os.environ.get('REVOKE_P_COLL', 0.000001)


# https://en.wikipedia.org/wiki/Bloom_filter
n_hash_func = -log(REVOKE_P_COLL)
n_bits_per_element = -1.44 * log(REVOKE_P_COLL)
n_bits = round((-REVOKE_MAX_TOKENS * log(REVOKE_P_COLL)) / (log(2) ** 2.0))
print("Req bytes:", n_bits / 8)
