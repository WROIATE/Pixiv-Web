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
        $("div.card img,.pic-info").each(function () {
            $(this).removeClass("active")
        })
        if ($(slide.opts.$orig).find("div.id[favour]").attr("favour") == "true") {
            $("[data-fancybox-like] span").addClass("like")
        } else {
            $("[data-fancybox-like] span").removeClass("like")
        }
        $(slide.opts.$orig).find("img").toggleClass("active");
        $(slide.opts.$orig).find(".pic-info").toggleClass("active");
        if ($(".sticky-nav").offset().top - $(document).scrollTop() == 0) {
            $(".sticky-nav").removeClass("navfadeIn")
            $(".sticky-nav").addClass("navfadeOut")
        }
    },
    afterClose: function (instance, slide) {
        $(slide.opts.$orig).find("img").removeClass("active");
        $(slide.opts.$orig).find(".pic-info").removeClass("active");

        $(".sticky-nav").removeClass("navfadeOut")
        $(".sticky-nav").addClass("navfadeIn")

    },
    caption: function (instance, item) {
        var caption = `<div class="m-auto" style="width:fit-content"><h5>` + $(this).find(".post-title").text() +
            `</h5><a href="https://www.pixiv.net/artworks/` + $(this).find(".id").attr("id") +
            `" target="_blank" id="` + $(this).find(".id").attr("id") +
            `"><h6>` + $(this).find(".id").text() +
            `</h6></a></div>`
        return caption
    },
    buttons: [
        'like',
        'thumbs',
        "slideShow",
        'close'
    ]

})

$.fancybox.defaults.btnTpl.like = `
<button data-fancybox-like class="fancybox-button fancybox-button--like" title="Favour">
 <span class='oi oi-heart iconic-heart-md text-white'></span>
</button>
`
$('body').on('click', '[data-fancybox-like]', function () {
    $("[data-fancybox-like] span").toggleClass("like")
    let id = $(".fancybox-caption__body a").attr("id")
    let like = $("div.id#" + id + "[favour]").attr("favour")
    if (like == "false") {
        $("div.id#" + id + "[favour]").attr("favour", "true")
    } else {
        $("div.id#" + id + "[favour]").attr("favour", "false")
    }
    $.ajax({
        url: window.location.origin,
        method: "POST",
        data: {
            id: "id=" + id,
            favour: $("div.id#" + id + "[favour]").attr("favour")
        },
        dataType: "json",
        success: function (data) {
            console.log(data)
        }
    })
});