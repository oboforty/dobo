from functools import wraps
from time import time


instances = {}


def Singleton(cls):
    def get_instance():
        global instances

        if cls not in instances:
            instance = cls()
            instances[cls] = instance
        return instances[cls]

    return get_instance


def async_cache(ttl):
    """
    Caches wrapped function's return value for given seconds
    :param ttl: time to live, in seconds
    """
    cache = {}

    def _decorator(fn):
        @wraps(fn)
        async def _wrapped(*args, **kwargs):
            now = time()
            item = cache.get('item')

            if item is None or (item[0] and item[0] + ttl < now):
                cache['item'] = now, await fn(*args, **kwargs)

            return cache['item'][1]

        return _wrapped
    return _decorator
