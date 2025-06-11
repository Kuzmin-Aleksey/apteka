const StatusOk = 200;
const StatusNotFound = 404;
const StatusUnauthorized = 401;

function FormatTime(tStr) {
    let t = new Date(tStr)
    return ("0" + t.getDate()).slice(-2) + "-" + ("0" + (t.getMonth() + 1)).slice(-2) + "-" +
        t.getFullYear() + " " + ("0" + t.getHours()).slice(-2) + ":" + ("0" + t.getMinutes()).slice(-2);
}


function withAuth(xhr) {
    let token = getCookie("token");
    if (token === "" || token === undefined) {
        OnUnauthorized("not found token in cookie");
    }
    xhr.setRequestHeader("Authorization", "Bearer " + token);
}

function withAuthUrl() {
    let token = getCookie("token");
    if (token === "" || token === undefined) {
        OnUnauthorized("not found token in cookie");
    }
    return "authorization=Bearer " + token
}


function deleteTokenFromCookie() {
    deleteCookie("token")
}


function getBookings() {
    let bookingsCookie = getCookie("bookings")
    if (bookingsCookie) {
        let bookings = JSON.parse(bookingsCookie)
        if (bookings) {
            return bookings
        }
    }

    return []
}

function addBooking(booking_id) {
    let oldBookings = getBookings()
    let bookings = [booking_id]

    bookings.push(...oldBookings)
    setCookie("bookings", JSON.stringify(bookings), {path: "/"});
}

function deleteBooking(booking_id) {
    let bookings = getBookings()
    bookings = bookings.filter(booking => booking !== booking_id)
    setCookie("bookings", JSON.stringify(bookings), {path: "/"});
}


function addAltImage(selector) {
    document.querySelectorAll(selector).forEach(el => {
        function listener() {
            el.removeEventListener("error", listener)
            el.setAttribute("src", "/static/img/no_product.webp")
        }

        el.addEventListener("error", listener)
    })
}

var deepEqual = function (x, y) {
    if (x === y) {
        return true;
    }
    else if ((typeof x == "object" && x != null) && (typeof y == "object" && y != null)) {
        if (Object.keys(x).length != Object.keys(y).length)
            return false;

        for (var prop in x) {
            if (y.hasOwnProperty(prop))
            {
                if (! deepEqual(x[prop], y[prop]))
                    return false;
            }
            else
                return false;
        }

        return true;
    }
    else
        return false;
}


let show = false

function loadingIndicator() {
    // Проверяем, существует ли уже индикатор
    let loader = document.getElementById('globalLoader');

    if (!show) {
        // Если индикатора нет - создаем его
        if (!loader) {
            loader = document.createElement('div');
            loader.id = 'globalLoader';
            loader.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background-color: rgba(0, 0, 0, 0.5);
                display: flex;
                justify-content: center;
                align-items: center;
                z-index: 9999;
            `;

            // Создаем элемент для спиннера
            const spinner = document.createElement('div');
            spinner.style.cssText = `
                width: 50px;
                height: 50px;
                border: 5px solid #f3f3f3;
                border-top: 5px solid #3498db;
                border-radius: 50%;
                animation: spin 1s linear infinite;
            `;

            // Добавляем анимацию
            const style = document.createElement('style');
            style.textContent = `
                @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
            `;

            loader.appendChild(spinner);
            document.head.appendChild(style);
            document.body.appendChild(loader);
        }
        loader.style.display = 'flex';
        show = true
    } else {
        if (loader) {
            loader.style.display = 'none';
        }
        show = false
    }
}

function hideModal(modal) {
    modal.querySelector("button[data-bs-dismiss=modal]").click()
}


function formatPrice(price, promo) {
    if (promo) {
        if (promo.is_percent) {
            price = price - price * promo.discount / 100
        } else {
            price -= promo.discount * 100
        }
    }

    price = Math.floor(price)

    let rub = Math.floor(price / 100);
    let kop = price % 100;

    let rubStr = "";

    let i = 0;

    rub.toString().split("").reverse().forEach(d => {
        i++
        rubStr =  d + rubStr
        if ((i % 3) === 0) {
            rubStr = " " + rubStr
        }
    })

    rubStr = rubStr.trimStart()

    if (kop === 0) {
        return rubStr
    }

    return `${rubStr},${kop}`

}


function getNoun(number, one, two, five) {
    let n = Math.abs(number);
    n %= 100;
    if (n >= 5 && n <= 20) {
        return five;
    }
    n %= 10;
    if (n === 1) {
        return one;
    }
    if (n >= 2 && n <= 4) {
        return two;
    }
    return five;
}