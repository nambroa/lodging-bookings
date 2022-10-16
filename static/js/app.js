function Prompt() {

    let toast = function (c) {
        const {
            msg = "",
            icon = "success",
            position = "top-end"
        } = c;

        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
            }
        })

        Toast.fire({})
    }

    let success = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c
        Swal.fire({
            icon: 'success',
            title: title,
            text: msg,
            footer: footer
        })
    }

    let error = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c
        Swal.fire({
            icon: 'error',
            title: title,
            text: msg,
            footer: footer
        })
    }

    // Modal open function. Name is custom due to the original Go course's lectures, but I will change it
    // in the future.
    async function custom(c) {
        const {
            msg = "",
            title = "",
            icon = "",
            showConfirmButton = true,
        } = c

        const {value: result} = await Swal.fire({
            title: title,
            html: msg,
            icon: icon,
            showConfirmButton: showConfirmButton,
            backdrop: false,
            showCancelButton: true,
            focusConfirm: false,
            willOpen: () => {
                if (c.willOpen !== undefined) {
                    c.willOpen()
                }
            },
            preConfirm: () => {
                return [
                    document.getElementById('start').value,
                    document.getElementById('end').value
                ]
            },
            didOpen: () => {
                if (c.didOpen !== undefined) {
                    c.didOpen()
                }
            }
        })

        if (result) {
            // If the user did not hit the cancel button in the modal. !== means EXACTLY.
            if (result.dismiss !== Swal.DismissReason.cancel) {
                if (result !== "") {
                    if (c.callback !== undefined) {
                        c.callback(result)
                    }
                } else {
                    c.callback(false);
                }
            } else {
                c.callback(false);
            }
        }
    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom
    }
}