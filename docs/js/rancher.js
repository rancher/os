jQuery(document).ready(function () {
  checkActiveState();

  // Set height
  var content_height = $(".col-sm-9").height();
  $(".col-sm-3").css("min-height", content_height + "px");

  $("ul.nav li a").click(function (event) {
    var link = $(this);
    var local_destination = link.attr("href");

    if ( link.data('toggle') )
    {
      return;
    }

    event.preventDefault();
    if ( window.location.protocol === 'file:' )
    {
      if ( local_destination.substr(-1) === '/' )
      {
        local_destination += 'index.html';
      }

      window.location.href = local_destination;
      return false;
    }

    $.get(local_destination, function (data) {
      var title = $(data).find("title").text();
      var local_content = $(data).find(".col-sm-9").html();
      changeUrl(title, local_destination);
      checkActiveState();
      $(".col-sm-9").html(local_content);
      //If has a class hash (jump on the page)
      if(link.hasClass('hash')){
        var local_destination_array = local_destination.split('#');
        var paragraph_id = local_destination_array[1];
        var paragraph_destination = $('.col-sm-9 #' + paragraph_id).offset().top;
        $('body,html').animate({scrollTop: paragraph_destination + "px"});
      }
      // Set height
      var content_height = $(".col-sm-9").height();
      $(".col-sm-3").css("min-height", content_height + "px");
    });

    setTimeout(function() {
      linkAnchors();
    },250);

  });
  linkAnchors();
});

function linkAnchors() {
  $('.content-container H3[id], .content-container H4[id], .content-container H5[id], .content-container H6[id]').each(function(idx, el) {
    $(el).append($('<a />').addClass('header-anchor').attr('href', '#' + el.id).html('<i class="fa fa-link" aria-hidden="true"></i>'));
  });
}

function changeUrl(title, url) {
    if (typeof (history.pushState) != "undefined") {
        var obj = {Title: title, Url: url};
        history.pushState(obj, obj.Title, obj.Url);
    }
}

function checkActiveState() {
  $('UL.nav UL').removeClass('in');
  $('UL A').each(function(idx, a) {
    var $a = $(a);
    if ( a.href === window.location.href )
    {
      $a.addClass('active');
      $a.closest('.list-group-submenu-submenu').parent().children('a').addClass("active");
      $a.closest('.list-group-submenu').parent().children('a').addClass("active");
      $a.parents('UL').addClass('in');
    }
    else
    {
      $a.removeClass('active');
    }
  });
}