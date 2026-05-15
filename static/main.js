window.onload = function() {
    console.log("Page loaded");
    updateData();
    setInterval(function() {
        updateData();
    }, 5000);
}

function updateData() {
    //console.log("Updating data...");

    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState === 4 && xhr.status === 200) {
            var data = JSON.parse(xhr.responseText);
            if (!data) {
                console.warn("No data received");
                return;
            }
            console.log("Data received:", data);

            document.getElementById("temperature").textContent = data.HaStateMap.Temperature.state + " " + data.HaStateMap.Temperature.attributes.unit_of_measurement;
            document.getElementById("humidity").textContent = data.HaStateMap.Humidity.state + " " + data.HaStateMap.Humidity.attributes.unit_of_measurement;
            document.getElementById("co2").textContent = data.HaStateMap.Co2.state + " " + data.HaStateMap.Co2.attributes.unit_of_measurement;


            document.getElementById("external-ip").textContent = data.HaStateMap.ExternalIp.state;
            document.getElementById("storage").textContent = data.DsmStorage.data.volumes[0].size.used + " / " + data.DsmStorage.data.volumes[0].size.total;
            
            var pcList = document.getElementById("pc-list");
            pcList.innerHTML = ""; // Clear previous list
            for (var i = 0; i < data.PcOnline.length; i++) {
                var pc = data.PcOnline[i];
                var element = document.createElement("h2");
                element.textContent = "🖥️ " + pc.Name + ": " + (pc.Online ? "Online 🟢" : "Offline 🔴");
                pcList.appendChild(element);
            }

            twemoji.parse(document.body);
        }
    };
    xhr.open("GET", "/data", true);
    xhr.send();
}