$.fancybox.defaults.btnTpl.like = `
<button data-fancybox-like class="fancybox-button fancybox-button--like" title="Favour">
 <span class='oi oi-heart iconic-heart-md'></span>
</button>
`
$.fancybox.defaults.btnTpl.download = `
<button data-fancybox-download class="fancybox-button fancybox-button--download" title="Download">
<span class="oi oi-cloud-download"></span>
</button>
`

$('body').on('click', '[data-fancybox-download]', function () {
    let id = $(".fancybox-caption__body a").attr("id")
    $.ajax({
        url: window.location.origin + "/" + "save",
        method: "POST",
        data: {
            id: id
        },
        success: function (data) {
            const $toast = $(`
            <div class="toast msgfade" role="alert" id="2" aria-live="assertive" aria-atomic="true">
            <div class="toast-header">
                <strong class="mr-auto">提示</strong>
                <button type="button" class="ml-2 mb-1 close" data-dismiss="toast" aria-label="Close">
                    <span aria-hidden="true">&times;</span>
                </button>
            </div>
            <div class="toast-body">
                已保存
            </div>
          </div>
            `);
            $toast.appendTo('.toastArea').toast({ delay: 3000 }).toast('show').on('hide.bs.toast', function () {

            });
        }
    })
});

$('body').on('click', '[data-fancybox-like]', function () {
    $("[data-fancybox-like] span").toggleClass("like")
    let id = $(".fancybox-caption__body a").attr("id")
    console.log(id)
    let like = $("img#" + id + "[favour]").attr("favour")
    if (like == "false") {
        $("img#" + id + "[favour]").attr("favour", "true")
    } else {
        $("img#" + id + "[favour]").attr("favour", "false")
    }
    $.ajax({
        url: window.location.origin,
        method: "POST",
        data: {
            id: "id=" + id,
            favour: $("img#" + id + "[favour]").attr("favour")
        },
        dataType: "json",
        success: function (data) {
            console.log(data)
        }
    })
});

$(document).scroll(function () {
    if ($("#menu").hasClass("menufadein")) {
        $("#menu").removeClass("menufadein")
        $("#menu").addClass("menufadeout")
        setTimeout(() => {
            $("#menu").css("display", "none")
        }, 200)
    }
});

$("div.sm-menu a").on("click", () => {
    if ($("#menu").hasClass("menufadein")) {
        $("#menu").removeClass("menufadein")
        $("#menu").addClass("menufadeout")
        setTimeout(() => {
            $("#menu").css("display", "none")
        }, 200)

    } else {
        $("#menu").removeClass("menufadeout")
        $("#menu").css("display", "block")
        $("#menu").addClass("menufadein")
    }
})

$('.form-control').keypress(function (e) {
    if (e.which == 13) {
        let id = $(this).val()
        console.log(id)
        window.location.href = "/search/" + id
    }
});

$("[data-fancybox=gallery]").fancybox({
    beforeShow: function (instance, slide) {
        $("div.card img,.pic-info").each(function () {
            $(this).removeClass("active")
        })
        if ($(slide.opts.$orig).find("img").attr("favour") == "true") {
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
        if ($("#menu").hasClass("menufadein")) {
            $("#menu").removeClass("menufadein")
            $("#menu").addClass("menufadeout")
            setTimeout(() => {
                $("#menu").css("display", "none")
            }, 200)
        }
    },
    afterClose: function (instance, slide) {
        $(slide.opts.$orig).find("img").removeClass("active");
        $(slide.opts.$orig).find(".pic-info").removeClass("active");

        $(".sticky-nav").removeClass("navfadeOut")
        $(".sticky-nav").addClass("navfadeIn")

    },
    caption: function (instance, item) {
        var caption = `<div class="m-auto" style="width:fit-content"><h5>` + $(this).find("img").attr("title") +
            `</h5><a href="https://www.pixiv.net/artworks/` + $(this).find("img").attr("id") +
            `" target="_blank" id="` + $(this).find("img").attr("id") +
            `"><h6>id=` + $(this).find("img").attr("id") +
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

$("[data-fancybox=gallery-result]").fancybox({
    beforeShow: function (instance, slide) {
        if ($(slide.opts.$orig).find("img").attr("favour") == "true") {
            $("[data-fancybox-like] span").addClass("like")
        } else {
            $("[data-fancybox-like] span").removeClass("like")
        }
        if ($(slide.opts.$orig).find("img").attr("local") == "true") {
            $("button[data-fancybox-download]").remove()
        }
        $(slide.opts.$orig).find("img").toggleClass("active");
        if ($(".sticky-nav").offset().top - $(document).scrollTop() == 0) {
            $(".sticky-nav").removeClass("navfadeIn")
            $(".sticky-nav").addClass("navfadeOut")
        }
    },
    afterClose: function (instance, slide) {
        $(slide.opts.$orig).find("img").removeClass("active");

        $(".sticky-nav").removeClass("navfadeOut")
        $(".sticky-nav").addClass("navfadeIn")

    },
    caption: function (instance, item) {
        var caption = `<div class="m-auto" style="width:fit-content"><h5>` + $(this).find("img").attr("title") +
            `</h5><a href="https://www.pixiv.net/artworks/` + $(this).find("img").attr("id") +
            `" target="_blank" id="` + $(this).find("img").attr("id") +
            `"><h6>id=` + $(this).find("img").attr("id") +
            `</h6></a></div>`
        return caption
    },
    buttons: [
        'download',
        'like',
        'thumbs',
        "slideShow",
        'close'
    ]
})

$("[data-fancybox=gallery-search]").fancybox({
    beforeShow: function (instance, slide) {
        $("div.card img,.pic-info").each(function () {
            $(this).removeClass("active")
        })
        if ($(slide.opts.$orig).find("img").attr("favour") == "true") {
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
        var caption = `<div class="m-auto" style="width:fit-content"><h5>` + $(this).find("img").attr("title") +
            `</h5><a href="https://www.pixiv.net/artworks/` + $(this).find("img").attr("id") +
            `" target="_blank" id="` + $(this).find("img").attr("id") +
            `"><h6>id=` + $(this).find("img").attr("id") +
            `</h6></a></div>`
        return caption
    },
    buttons: [
        'download',
        'like',
        'thumbs',
        "slideShow",
        'close'
    ]

})

$("#clean").on("click", () => {
    $.ajax({
        url: window.location.origin + "/" + "clean",
        method: "POST",
        data: {

        },
        success: () => {
            window.location.href = "/search/"
        }
    })
})


$("figure").each(function () {
    var img_url = $(this).find("img").data("src")
    var img = new Image()
    img.src = img_url
    check = () => {
        if (img.width > 0 && img.height > 0) {
            $(this).css("flex-grow", img.width * 100 / img.height)
            $(this).css("width", img.width * 100 / img.height)
            $(this).find("img").attr("src", img_url)
        }
    }
    // var set = setInterval(check, 40);
    img.onload = () => {
        $(this).css("flex-grow", img.width * 100 / img.height)
        $(this).css("width", img.width * 100 / img.height)
        $(this).find("img").attr("src", img_url)
    };

})


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


$(window).scroll(function () {

    var top1 = $(this).scrollTop();     //获取相对滚动条顶部的偏移

    if (top1 > 0) {      //当偏移大于200px时让图标淡入（css设置为隐藏）

        $(".up").fadeIn("fast");

    } else {
        //当偏移小于200px时让图标淡出
        $(".up").fadeOut("fast");
    }
});

//去往顶部
$(".up").click(function () {
    var t = 300
    if ($(window).scrollTop() / 10 > 300) {
        t = $(window).scrollTop() / 10
    }
    $("body , html").animate({ scrollTop: 0 }, t);   //300是所用时间
});

