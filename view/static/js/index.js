{
    let h
    $('body').on('show.bs.modal', ".modal", () => {
        h = $(window).scrollTop();
        $("body").css({
            "top": -h
        });
    }).on("hidden.bs.modal", ".modal", () => {
        $(window).scrollTop(h);
    });
}

$("body").on("click", "a#reload", () => {

    var ws = new WebSocket("ws://" + window.location.host + "/reload" + $(".tag a.active").attr("href"))
    ws.onmessage = function (evt) {
        let data = JSON.parse(evt.data)
        if (data["total"] != 0) {
            let per = (data["total"] - data["num"]) / data["total"]
            $("div[role=progressbar]").css("width", per * 100 + "%")
        }
    }
    ws.onclose = function (evt) {
        console.log("Connection closed.")
        $("div[role=progressbar]").css("width", "100%")
        setTimeout(window.location.reload.bind(window.location), 1000)
    }
})

$("[data-fancybox]").fancybox({
    beforeShow: function (instance, slide) {
        $("a.card img,.pic-info").each(function () {
            $(this).removeClass("active")
        })
        $(slide.opts.$orig).find("img").toggleClass("active");
        $(slide.opts.$orig).find(".pic-info").toggleClass("active");
    },
    afterClose: function (instance, slide) {
        $(slide.opts.$orig).find("img").removeClass("active");
        $(slide.opts.$orig).find(".pic-info").removeClass("active");
    },
})
