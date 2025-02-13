from setuptools import setup, find_packages

# with open("README.md", "r", encoding="utf-8") as fh:
#     long_description = fh.read()

setup(
    name="dobodb",
    version="0.1.0",
    author="Rajmund Csombordi",
    author_email="rajmund.csombordi@gmail.com",
    description="Write-optimized AP database",
    # long_description=long_description,
    # long_description_content_type="text/markdown",
    # packages=find_packages(),
    install_requires=[
        # "requests",
    ],
    classifiers=[
        "Programming Language :: Python :: 3",
        "Operating System :: OS Independent",
    ],
    python_requires='>=3.10',
)
