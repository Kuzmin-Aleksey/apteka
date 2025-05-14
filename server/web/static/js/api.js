




function DeleteStore(storeId) {

}



function CheckToken() {
    let xhr = new XMLHttpRequest();
    xhr.responseType = "json"
    xhr.onload = () => {
        CheckCode(xhr)
    }
    xhr.open("GET", "/api/auth/check-token")
    withAuth(xhr)
    xhr.send()
}

function GetPromotions(callback) {
    let xhr = new XMLHttpRequest();
    xhr.responseType = "json"
    xhr.onload = () => {
        if (!CheckCode(xhr)) {
            return
        }
        callback(xhr.response)
    }


    xhr.open("GET", `/api/promotion/get`)
    xhr.send()
}

function CreatePromotion(promotion) {
    let xhr = new XMLHttpRequest();

    xhr.responseType = "json"
    xhr.onload = () => {
        CheckCode(xhr)
    }

    xhr.open("POST", `/api/promotion/new`)
    withAuth(xhr)
    console.log(promotion)
    xhr.send(promotion)
}

function UploadPromotionsDoc(file, onprogress) {

}

function UpdatePromotion(promotion) {
    let xhr = new XMLHttpRequest();

    xhr.responseType = "json"
    xhr.onload = () => {
        CheckCode(xhr)
    }

    xhr.open("POST", `/api/promotion/update`)
    withAuth(xhr)
    xhr.send(JSON.stringify(promotion))
}


function DeletePromotions() {
    let xhr = new XMLHttpRequest();

    xhr.responseType = "json"
    xhr.onload = () => {
        CheckCode(xhr)
    }


    xhr.open("POST", `/api/promotion/delete-all`)
    withAuth(xhr)
    xhr.send()
}

function SetBookingDelay(delay) {
    let xhr = new XMLHttpRequest();

    xhr.responseType = "json"
    xhr.onload = () => {
        CheckCode(xhr)
    }

    xhr.open("POST", `/api/booking/set-delay?delay=${delay}`)
    withAuth(xhr)
    xhr.send()
}

function LoadImages() {
    let xhr = new XMLHttpRequest();

    xhr.responseType = "json"
    xhr.onload = () => {
        if (!CheckCode(xhr)) {
            return
        }
    }


    xhr.open("POST", `/api/images/load`)
    withAuth(xhr)
    xhr.send()
}







