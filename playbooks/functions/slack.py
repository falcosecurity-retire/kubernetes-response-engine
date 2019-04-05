import os
import playbooks
from playbooks import infrastructure


playbook = playbooks.AddMessageToSlack(
    infrastructure.SlackClient(os.environ['SLACK_WEBHOOK_URL'])
)


def handler(event, context):
    alert = playbooks.falco_alert(event)
    if alert:
        playbook.run(alert)
