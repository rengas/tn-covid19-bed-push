$(document).ready(function(){
    // Initialize the default app
    // Your web app's Firebase configuration
    const firebaseConfig = {
        apiKey: "AIzaSyB_yoK3801boAiw6jD_wHCA19N-FtH8Pvs",
        authDomain: "tn-covid-bed-alert.firebaseapp.com",
        projectId: "tn-covid-bed-alert",
        storageBucket: "tn-covid-bed-alert.appspot.com",
        messagingSenderId: "377040005391",
        appId: "1:377040005391:web:1ebc106714b848cc1fb67a"
    };
    // Initialize Firebase

    firebase.initializeApp(firebaseConfig)

    const messaging = firebase.messaging();

    navigator.serviceWorker.register('./static/firebase-messaging-sw.js')
        .then((registration) => {
            messaging.useServiceWorker(registration);
        });

    messaging.onMessage(function(payload) {
        console.log('[main-sw.js] Received forground message ', payload);
        const notificationTitle = payload.notification.title;
        const notificationOptions = {
            body: payload.notification.body,
            icon: payload.notification.icon,
        };

        if (!("Notification" in window)) {
            console.log("This browser does not support system notifications.");
        } else if (Notification.permission === "granted") {
            // If it's okay let's create a notification
            var notification = new Notification(notificationTitle,notificationOptions);
            notification.onclick = function(event) {
                event.preventDefault();
                window.open(payload.notification.click_action , '_blank');
                notification.close();
            }
        }
    });

    $("#myButton").click(function() {
        console.log('Requesting permission...');
        Notification.requestPermission().then((permission) => {
            if (permission === 'granted') {
                messaging.getToken({vapidKey: 'BK56Jk0KnKiHP6VZ7AMk0I03ztNptBq2srJGE0NrK7LBQbHBtw-DnWJ1feX3XHt90NOXsfqYSH00WoCEtfBQhEg'}).then((currentToken) => {
                    if (currentToken) {
                        var data = $('#submitForm').serializeArray();
                        data.push({
                            name:'token',
                            value: currentToken
                        })
                        $.ajax({
                            type: 'POST',
                            url: 'subscribe',
                            data: data,
                            dataType: 'json'
                        }).done(function(data) {
                            alert("Thanks for the submission!");
                            $('#submitForm')[0].reset();
                        });

                    } else {
                        console.log('No registration token available. Request permission to generate one.');
                    }
                }).catch((err) => {
                    console.log('An error occurred while retrieving token. ', err);
                });
            } else {
                console.log('Unable to get permission to notify.');
            }
        });
    });

    $('.group').hide();
    $("select").change(function () {
        $('.group').hide();
        $('#'+$(this).val()).show();
    })
});