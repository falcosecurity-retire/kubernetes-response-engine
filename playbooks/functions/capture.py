import sys
import base64
import os.path

sys.path.append(os.path.join(os.path.abspath(os.path.dirname(__file__))))

import os
import playbooks
from playbooks import infrastructure

if "AWS_S3_BUCKET" in os.environ and "AWS_ACCESS_KEY_ID" in os.environ and "AWS_SECRET_ACCESS_KEY" in os.environ:
    playbook = playbooks.StartSysdigCaptureForContainerS3(
        infrastructure.KubernetesClient(),
        int(os.environ.get('CAPTURE_DURATION', 120)),
        os.environ['AWS_S3_BUCKET'],
        os.environ['AWS_ACCESS_KEY_ID'],
        os.environ['AWS_SECRET_ACCESS_KEY']
    )

if "GCLOUD_BUCKET" in os.environ:
    playbook = playbooks.StartSysdigCaptureForContainerGcloud(
        infrastructure.KubernetesClient(),
        int(os.environ.get('CAPTURE_DURATION', 120)),
        os.environ['GCLOUD_BUCKET']
    )


def handler(event, context):
    playbook.run(playbooks.falco_alert(event))
