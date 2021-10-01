function startTimer(startTime, timerElement) {
    let i = setInterval(function() {
        let now = new Date().getTime();
        let duration = startTime - now;
        let days = Math.floor(duration / (1000 * 60 * 60 * 24));
        let hours = Math.floor((duration % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        let minutes = Math.floor((duration % (1000 * 60 * 60)) / (1000 * 60));
        let seconds = Math.floor((duration % (1000 * 60)) / 1000);
        let daysStr = days + "d ";
        let hoursStr = hours + "h ";
        let minutesStr = minutes + "m "
        let secondsStr = seconds + "s";

        if(days < 10) {
            if(days == 0)
                daysStr = "";
            else
                daysStr = "0" + days + "d ";
        }
        if(hours < 10)
            hoursStr = "0" + hours + "h ";
        if(minutes < 10)
            minutesStr = "0" + minutes + "m ";
        if(seconds < 10)
            secondsStr = "0" + seconds + "s";
        timerElement.textContent = daysStr + hoursStr + minutesStr + secondsStr;
        if (duration < 0) {
            clearInterval(i);
            timerElement.textContent = "Stay Tuned";
        }
    }, 1000);
}
