import playbooks
from playbooks import infrastructure


playbook = playbooks.DeletePod(
    infrastructure.KubernetesClient()
)

def handler(event, context):
    alert = playbooks.falco_alert(event)
    if alert:
        playbook.run(alert)
