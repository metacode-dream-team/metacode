<#-- my-theme/login/login.ftl -->
<#noparse>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Redirecting...</title>

    <script type="text/javascript">
        (function() {
            const frontendUrl = 'http://localhost:3000/oauth/error';
            const params = new URLSearchParams(window.location.search);
            let redirectUrl = frontendUrl;

            if (params.has('error')) {
                const error = encodeURIComponent(params.get('error'));
                const description = encodeURIComponent(params.get('error_description') || '');
                redirectUrl = frontendUrl + '?error=' + error + '&description=' + description;
            }

            window.location.href = redirectUrl;
        })();
    </script>
</head>
<body>
    <p>Redirecting...</p>
</body>
</html>
</#noparse>