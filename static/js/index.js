$('#myModal').on('shown.bs.modal', function () { $('#myInput').trigger('focus') })
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

$(function () { $('[data-toggle="tooltip"]').tooltip() })