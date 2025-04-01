## ActiveDirectory Terraform example

The files in this directory provide an example of a maliciously re-configured ActiveDirectory service.

```
╰─$ diff secure.tf insecure.tf
31,32c31,32
<       access_token_issuance_enabled = false
<       id_token_issuance_enabled     = false
---
>       access_token_issuance_enabled = true
>       id_token_issuance_enabled     = true
45c45
<   sign_in_audience = "AzureADMyOrg"
---
>   sign_in_audience = "AzureADandPersonalMicrosoftAccount"
47c47
<   prevent_duplicate_names = true
---
>   prevent_duplicate_names = false
```

We have a workflow to show signing provenance for `secure.tf`, and in a demonstration we will show replacing it with the contents of `insecure.tf`. 

While admittedly somewhat contrived, this gives us a way to show the tamper-proof nature of an attestation, proving that this can be a mitigation against Monster-in-the-Middle attacks in configuration pipelines.