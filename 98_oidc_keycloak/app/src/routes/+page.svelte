<script lang="ts">
import Keycloak from 'keycloak-js';

const handleClick = async() => {
    console.log("handleClick function invoked")
    const keycloak = new Keycloak({
        url: 'http://localhost:8080',
        realm: 'ossc',
        clientId: 'spa1'
    });
    keycloak.init({
        redirectUri: 'http://localhost:8000'
    })
    .then((authenticated) => {
        console.log("authenticated.")
    })
};

</script>


<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p>


<button on:click={handleClick}>login</button>