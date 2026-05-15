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
            
            // pveStats.innerHTML = ""; // Clear previous stats
            var totalMem = 0;
            var totalMemUsed = 0;
            var totalCpu = 0;
            var totalCpuUsed = 0;
            for (var i = 0; i < data.PveMetrics.data.length; i++) {
                var node = data.PveMetrics.data[i];
                totalMem += node.maxmem;
                totalMemUsed += node.mem;

                totalCpu += node.maxcpu;
                totalCpuUsed += node.cpu;
            }
            // console.log("Total Memory:", totalMem, "Used Memory:", totalMemUsed);
            // console.log("Total CPU:", totalCpu, "Used CPU:", totalCpuUsed, "CPU Usage:", ((totalCpuUsed / totalCpu) * 100).toFixed(2) + "%");

            var pveStats = document.getElementById("pve-stats");
            pveStats.textContent = "Mem: " + ((totalMemUsed / totalMem) * 100).toFixed(2) + "%" + " CPU: " + ((totalCpuUsed / totalCpu) * 100).toFixed(2) + "%";

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