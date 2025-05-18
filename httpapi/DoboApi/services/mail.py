import boto3

from DoboApi.settings import settings

ses = boto3.client('ses', region_name='eu-central-1')


class MailManager:
    plain_verify_mail = False

    def send_verify_email(self, email_address: str):
        if self.plain_verify_mail:
            resp = ses.verify_email_identity(
                EmailAddress=email_address
            )
        else:
            resp = ses.send_custom_verification_email(
                EmailAddress=email_address,
                TemplateName=settings.verify_email_template,
                # ConfigurationSetName='string'
            )

        return bool(resp)
