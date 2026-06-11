from gke_mcp.tools.params import (
    Project,
    LocationRequired,
    LocationOptional,
    Cluster,
    NodePool,
    Operation
)

def test_project():
    p = Project("test-proj")
    assert p.project_id_path == "projects/test-proj"

def test_location_required():
    l = LocationRequired("test-proj", "us-central1")
    assert l.location_path == "projects/test-proj/locations/us-central1"

def test_location_optional():
    l1 = LocationOptional("test-proj")
    assert l1.location_path == "projects/test-proj/locations/-"
    
    l2 = LocationOptional("test-proj", "us-east1")
    assert l2.location_path == "projects/test-proj/locations/us-east1"

def test_cluster():
    c = Cluster("test-proj", "us-central1", "my-cluster")
    assert c.cluster_path == "projects/test-proj/locations/us-central1/clusters/my-cluster"

def test_node_pool():
    np = NodePool("test-proj", "us-central1", "my-cluster", "my-pool")
    assert np.node_pool_path == "projects/test-proj/locations/us-central1/clusters/my-cluster/nodePools/my-pool"

def test_operation():
    op = Operation("test-proj", "us-central1", "op-123")
    assert op.operation_path == "projects/test-proj/locations/us-central1/operations/op-123"
