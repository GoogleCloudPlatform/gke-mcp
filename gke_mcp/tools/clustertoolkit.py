import os
import subprocess
import logging
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.clustertoolkit")

def cluster_toolkit_download(cfg: Config, download_directory: str) -> str:
    """Cluster Toolkit, is open-source software offered by Google Cloud which simplifies the process for you to create Google Kubernetes Engine clusters and deploy high performance computing (HPC), artificial intelligence (AI), and machine learning (ML). It is designed to be highly customizable and extensible, and intends to address the deployment needs of a broad range of use cases. This tool will download the public git repository so that Cluster Toolkit can be used."""
    if not download_directory:
        raise ValueError("download_directory argument cannot be empty")
        
    download_dir = download_directory
    if not download_dir.endswith("cluster-toolkit"):
        download_dir = os.path.join(download_dir, "cluster-toolkit")
        
    cmd = ["git", "clone", "https://github.com/GoogleCloudPlatform/cluster-toolkit.git", download_dir]
    logger.info(f"Downloading Cluster Toolkit: {' '.join(cmd)}")
    
    try:
        res = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return res.stdout if res.stdout else "Successfully cloned Cluster Toolkit."
    except subprocess.CalledProcessError as e:
        raise RuntimeError(f"failed to clone Cluster Toolkit: {e.stdout} {e.stderr}")
