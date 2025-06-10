function getStoreId() {
    let storeIdCookie = getCookie("store")
    if (storeIdCookie) {
        let storeId = Number.parseInt(storeIdCookie)
        return storeId ? storeId : null
    }
    return null
}

function saveStoreId(storeId) {
    setCookie("store", storeId, {path: "/"})
}


const params = new URLSearchParams(location.search);
let storeId = Number.parseInt(params.get("store"));
let storeInfo = null;
let stores = [];

if (storeId) {
    saveStoreId(storeId);
} else {
    storeId = getStoreId()
}

function CheckCode(xhr) {
    if (xhr.status === StatusOk) {
        return true
    }
    console.error(xhr.response)
    alert(`Ошибка ${xhr.status}`)
    return false
}


// Получение информации об аптеке
function getStoreInfo(callback) {
    const wg = new WaitGroup();
    wg.add(1)

    let xhr = new XMLHttpRequest();

    //xhr.responseType = "json"
    xhr.onload = () => {
        wg.done()
        if (!CheckCode(xhr)) {
            return;
        }

        stores = JSON.parse(xhr.response);

        if (storeId) {
            stores.forEach(s => {
                if (s.id === storeId) {
                    storeInfo = s;
                }
            })
        }
    }
    xhr.open("GET", `/api/stores/get`, false);
    xhr.send();

    wg.wait();
}



getStoreInfo()