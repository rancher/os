$(document).ready(function(){
  var $menu = $("#menu");

  $menu.slicknav({
    label: "Navigation",
    duration: 250,
    appendTo: "header"
  });

  // Expand the tree to the current URL
  var found = false;
  $menu.find("a").each(function(idx, link) {
    if ( link.href === window.location.href) {
      expand(link);
      $(link).addClass('active');
      found = true;
    }
  });

  // Expand items when clicking on the tree
  $menu.on("click", "li a", function(event) {
    var open = isExpanded(this);
    collapseAll();
    if ( !open ) {
      expand(this);
    }

    // Don't change the URL for links that are just an anchor like "#hosts"
    if ( $(this).attr('href').substr(0,1) === '#' ) {
      event.preventDefault();
      event.stopPropagation();
    }
  });

  $('.content-container').on('mouseenter', 'h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]', function(e) {
    $(e.target).append($('<a />').addClass('header-anchor').attr('href', '#' + e.target.id).html('<i class="fa fa-link " aria-hidden="true"></i>'));
  });

  $('.content-container').on('mouseleave', 'h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]', function(e) {
    $(e.target).parent().find('.header-anchor').remove();
  });
});

function isExpanded(link) {
  return $(link).closest("li").hasClass("active");
}

function expand(link) {
  $(link).parentsUntil("#menu","li").addClass("active");
}

function collapseAll() {
  $('#menu .active').removeClass('active');
}
