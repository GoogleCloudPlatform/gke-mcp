from typing import Optional

class Project:
    def __init__(self, project_id: str):
        self.project_id = project_id

    @property
    def project_id_path(self) -> str:
        return f"projects/{self.project_id}"

class LocationRequired(Project):
    def __init__(self, project_id: str, location: str):
        super().__init__(project_id)
        self.location = location

    @property
    def location_path(self) -> str:
        return f"{self.project_id_path}/locations/{self.location}"

class LocationOptional(Project):
    def __init__(self, project_id: str, location: Optional[str] = None):
        super().__init__(project_id)
        self.location = location

    @property
    def location_path(self) -> str:
        loc = self.location if self.location else "-"
        return f"{self.project_id_path}/locations/{loc}"

class Cluster(LocationRequired):
    def __init__(self, project_id: str, location: str, cluster_name: str):
        super().__init__(project_id, location)
        self.cluster_name = cluster_name

    @property
    def cluster_path(self) -> str:
        return f"{self.location_path}/clusters/{self.cluster_name}"

class NodePool(Cluster):
    def __init__(self, project_id: str, location: str, cluster_name: str, node_pool_name: str):
        super().__init__(project_id, location, cluster_name)
        self.node_pool_name = node_pool_name

    @property
    def node_pool_path(self) -> str:
        return f"{self.cluster_path}/nodePools/{self.node_pool_name}"

class Operation(LocationRequired):
    def __init__(self, project_id: str, location: str, operation_id: str):
        super().__init__(project_id, location)
        self.operation_id = operation_id

    @property
    def operation_path(self) -> str:
        return f"{self.location_path}/operations/{self.operation_id}"
